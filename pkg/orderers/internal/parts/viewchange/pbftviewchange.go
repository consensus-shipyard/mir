package viewchange

import (
	"bytes"

	es "github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	common2 "github.com/filecoin-project/mir/pkg/orderers/common"
	"github.com/filecoin-project/mir/pkg/orderers/internal/common"
	"github.com/filecoin-project/mir/pkg/orderers/internal/parts/goodcase"
	ot "github.com/filecoin-project/mir/pkg/orderers/types"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"
	pbftpbdsl "github.com/filecoin-project/mir/pkg/pb/pbftpb/dsl"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	timertypes "github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// ============================================================
// OrdererEvent handling
// ============================================================

func IncludeViewChange( //nolint:gocognit
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	logger logging.Logger,
) {
	// UponSignResult with a pftpbtypes.ViewChange context, process a newly generated signature of a ViewChange message.
	// Then, send  a SignedViewChange message to the new leader (PBFT primary)
	cryptopbdsl.UponSignResult(m, func(signature []byte, context *pbftpbtypes.ViewChange) error {

		// Compute the primary of this view (using round-robin on the membership)
		primary := state.Segment.PrimaryNode(context.View)

		// Repeatedly send the ViewChange message. Repeat until this instance of PBFT is garbage-collected,
		// i.e., from the point of view of the PBFT protocol, effectively forever.

		eventpbdsl.TimerRepeat(m, moduleConfig.Timer,
			[]*eventpbtypes.Event{transportpbevents.SendMessage(
				moduleConfig.Net,
				pbftpbmsgs.SignedViewChange(moduleConfig.Self, context, signature),
				[]t.NodeID{primary},
			)},
			timertypes.Duration(params.Config.ViewChangeResendPeriod),
			tt.RetentionIndex(params.Config.EpochNr),
		)
		return nil
	})

	cryptopbdsl.UponSigVerified(m, func(nodeID t.NodeID, error error, context *pbftpbtypes.SignedViewChange) error {
		// Ignore events with invalid signatures.
		if error != nil {
			logger.Log(logging.LevelWarn,
				"Ignoring invalid signature, ignoring event.",
				"from", nodeID,
				"error", error,
			)
			return nil
		}
		// Verify all signers are part of the membership
		if !sliceutil.Contains(params.Config.Membership, nodeID) {
			logger.Log(logging.LevelWarn,
				"Ignoring SigVerified message as it contains signatures from non members, ignoring event (with all signatures).",
				"from", nodeID,
			)
			return nil
		}

		applyVerifiedViewChange(m, state, params, moduleConfig, context, nodeID, logger)
		return nil
	})

	cryptopbdsl.UponSigsVerified(m, func(nodeIds []t.NodeID, errors []error, allOk bool, context *pbftpbtypes.NewView) error {
		// Ignore events with invalid signatures.
		if !allOk {
			logger.Log(logging.LevelWarn,
				"Ignoring invalid signature, ignoring event (with all signatures).",
				"from", nodeIds,
				"errors", errors,
			)
			return nil
		}

		// Verify all signers are part of the membership
		if !sliceutil.ContainsAll(params.Config.Membership, nodeIds) {
			logger.Log(logging.LevelWarn,
				"Ignoring SigsVerified message as it contains signatures from non members, ignoring event (with all signatures).",
				"from", nodeIds,
			)
			return nil
		}
		// Serialize obtained Preprepare messages for hashing.
		dataToHash := make([]*hasherpbtypes.HashData, len(context.Preprepares))
		for i, preprepare := range context.Preprepares { // Preprepares in a NewView message are sorted by sequence number.
			dataToHash[i] = common.SerializePreprepareForHashing(preprepare)
		}

		// Request hashes of the Preprepare messages.
		hasherpbdsl.Request(
			m,
			moduleConfig.Hasher,
			dataToHash,
			context,
		)
		return nil
	})

	// UponResult for an empty preprepare (the context is just the view number)
	hasherpbdsl.UponResult(m, func(digests [][]byte, context *ot.ViewNr) error {
		// Convenience variable.
		view := *context
		// Ignore hash result if the view has advanced in the meantime.
		if view < state.View {
			logger.Log(logging.LevelDebug, "Aborting construction of NewView after hashing. View already advanced.",
				"hashView", view,
				"localView", state.View,
			)
			return nil
		}

		// Look up the corresponding view change state.
		// No presence check needed, as the entry must exist.
		vcstate := state.ViewChangeStates[view]

		// Set the digests of empty Preprepares that have just been computed.
		if err := vcstate.SetEmptyPreprepareDigests(digests); err != nil {
			return es.Errorf("error setting empty preprepare digests: %w", err)
		}

		// Check if all preprepare messages that need to be re-proposed are locally present.
		vcstate.SetLocalPreprepares(state, view)
		if vcstate.HasAllPreprepares() {
			logger.Log(logging.LevelDebug, "All Preprepares present, sending NewView.")
			// If we have all Preprepares, start view change.
			sendNewView(m, state, moduleConfig, view, vcstate, logger)
			return nil
		}

		// If some Preprepares for re-proposing are still missing, fetch them from other nodes.
		logger.Log(logging.LevelDebug, "Some Preprepares missing. Asking for retransmission.")
		askForMissingPreprepares(m, moduleConfig, vcstate)
		return nil
	})

	// UponResultOne with a MissingPreprepare as context
	hasherpbdsl.UponResultOne(m, func(digest []byte, context *pbftpbtypes.MissingPreprepare) error {
		// Convenience variable
		sn := context.Preprepare.Sn

		// Look up the latest (with the highest) pending view change state.
		// (Only the view change states that might be waiting for a missing preprepare are considered.)
		vcstate, view := latestPendingVCState(state)

		// Ignore preprepare if received in the meantime or if view has already advanced.
		// (Such a situation can occur if missing Preprepares arrive late.)
		if pp, ok := vcstate.Preprepares[sn]; (ok && pp != nil) || view < state.View {
			return nil
		}

		// Add the missing preprepare message if its digest matches, updating its view.
		// Note that copying a preprepare with an updated view preserves its hash.
		if bytes.Equal(vcstate.Reproposals[sn], digest) && vcstate.Preprepares[sn] == nil {
			preprepareCopy := *context.Preprepare
			preprepareCopy.View = view
			vcstate.Preprepares[sn] = &preprepareCopy
		}

		logger.Log(logging.LevelDebug, "Received missing Preprepare message.", "sn", sn)

		// If this was the last missing preprepare message, proceed to sending a NewView message.
		if vcstate.HasAllPreprepares() {
			sendNewView(m, state, moduleConfig, view, vcstate, logger)
		}
		return nil
	})

	// UponResult with a NewView as context
	hasherpbdsl.UponResult(m, func(digests [][]byte, context *pbftpbtypes.NewView) error {
		// Ignore message if old.
		if context.View < state.View {
			return nil
		}

		// Create a temporary view change state object
		// to use for reconstructing the re-proposals from the obtained view change messages.
		vcState := common.NewPbftViewChangeState(state.Segment.SeqNrs(), state.Segment.Membership)

		// Feed all obtained ViewChange messages to the view change state.
		for i, signedViewChange := range context.SignedViewChanges {
			vcState.AddSignedViewChange(signedViewChange, t.NodeID(context.ViewChangeSenders[i]), logger)
		}

		// If the obtained ViewChange messages are not sufficient to infer all re-proposals, ignore NewView message.
		if !vcState.EnoughViewChanges() {
			return nil
		}

		// Verify if the re-proposed hashes match the obtained Preprepares.
		i := 0
		prepreparesMatching := true
		maputil.IterateSorted(vcState.Reproposals, func(sn tt.SeqNr, digest []byte) (cont bool) {

			// If the expected digest is empty, it means that the corresponding Preprepare is an "aborted" one.
			// In this case, check the Preprepare directly.
			if len(digest) == 0 {
				prepreparesMatching = validFreshPreprepare(context.Preprepares[i], context.View, sn)
			} else {
				prepreparesMatching = bytes.Equal(digest, digests[i])
			}

			i++
			return prepreparesMatching
		})

		// If the NewView contains mismatching Preprepares, ignore the message.
		if !prepreparesMatching {
			logger.Log(logging.LevelWarn, "Hash mismatch in received NewView. Ignoring.", "view", context.View)
			return nil
		}

		// If all the checks passed, (TODO: make sure all the checks of the NewView message have been performed!)
		// enter the new view and clear the InViewChange flag, as we continue normal operation now.
		//This call is necessary if this node was not among those initiating the view change
		err := state.InitView(m, params, moduleConfig, context.View, logger)
		if err != nil {
			return err
		}
		state.InViewChange = false

		// Apply all the Preprepares contained in the NewView.
		// This needs to be applied before processing the buffered messages,
		// in case a malicious node has sent conflicting ones before.
		primary := state.Segment.PrimaryNode(context.View)
		for _, preprepare := range context.Preprepares {
			goodcase.ApplyMsgPreprepare(m, state, params, moduleConfig, preprepare, primary, logger)
		}

		// Apply all messages buffered for this view.
		for from, msgBuf := range state.MessageBuffers {
			msgBuf.Iterate(func(source t.NodeID, msgPb proto.Message) messagebuffer.Applicable {
				msgView, err := getMsgView(msgPb)
				if err != nil {
					return messagebuffer.Invalid
				} else if msgView < state.View {
					return messagebuffer.Past
				} else if msgView == state.View {
					return messagebuffer.Current
				} else {
					return messagebuffer.Future
				}
			}, func(source t.NodeID, msgPb proto.Message) {
				goodcase.ApplyBufferedMsg(m, state, params, moduleConfig, msgPb, from, logger)
			})
		}

		return nil
	})

	pbftpbdsl.UponViewChangeSegTimeout(m, func(viewChangeSegTimeout uint64) error {
		return applyViewChangeSegmentTimeout(m, state, params, moduleConfig, ot.ViewNr(viewChangeSegTimeout), logger)
	})

	pbftpbdsl.UponViewChangeSNTimeout(m, func(view ot.ViewNr, numCommitted uint64) error {
		return applyViewChangeSNTimeout(m, state, params, moduleConfig, view, numCommitted, logger)
	})

	pbftpbdsl.UponSignedViewChangeReceived(m, func(from t.NodeID, viewChange *pbftpbtypes.ViewChange, signature []byte) error {
		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		svc := &pbftpbtypes.SignedViewChange{
			ViewChange: viewChange,
			Signature:  signature,
		}
		applyMsgSignedViewChange(m, moduleConfig, svc, from)
		return nil
	})

	pbftpbdsl.UponPreprepareRequestReceived(m, func(from t.NodeID, digest []byte, sn tt.SeqNr) error {
		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}

		applyMsgPreprepareRequest(m, state, moduleConfig, digest, sn, from)
		return nil
	})

	pbftpbdsl.UponMissingPreprepareReceived(m, func(from t.NodeID, preprepare *pbftpbtypes.Preprepare) error {
		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		applyMsgMissingPreprepare(m, state, moduleConfig, preprepare)
		return nil
	})

	pbftpbdsl.UponNewViewReceived(m, func(
		from t.NodeID,
		view ot.ViewNr,
		viewChangeSenders []string,
		signedViewChanges []*pbftpbtypes.SignedViewChange,
		preprepareSeqNrs []tt.SeqNr,
		preprepares []*pbftpbtypes.Preprepare,
	) error {

		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		nv := &pbftpbtypes.NewView{
			View:              view,
			ViewChangeSenders: viewChangeSenders,
			SignedViewChanges: signedViewChanges,
			PreprepareSeqNrs:  preprepareSeqNrs,
			Preprepares:       preprepares,
		}
		applyMsgNewView(m, state, moduleConfig, nv, from)
		return nil
	})
}

// applyViewChangeSNTimeout applies the view change SN timeout event
// triggered some time after a value is committed.
// If nothing has been committed since, triggers a view change.
func applyViewChangeSNTimeout(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	view ot.ViewNr,
	numCommitted uint64,
	logger logging.Logger,
) error {

	// If the view is still the same as when the timer was set up,
	// if nothing has been committed since then, and if the segment-level checkpoint is not yet stable
	if view == state.View &&
		int(numCommitted) == state.NumCommitted(state.View) &&
		!state.SegmentCheckpoint.Stable(state.Segment.Membership) {

		// Start the view change sub-protocol.
		logger.Log(logging.LevelWarn, "View change SN timer expired.",
			"view", state.View, "numCommitted", numCommitted)
		return startViewChange(m, state, params, moduleConfig, logger)
	}

	// Do nothing otherwise.
	return nil
}

// applyViewChangeSegmentTimeout applies the view change segment timeout event
// triggered some time after a segment is initialized.
// If not all slots have been committed, and the view has not advanced, triggers a view change.
func applyViewChangeSegmentTimeout(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	view ot.ViewNr,
	logger logging.Logger,
) error {

	// TODO: All slots being committed is not sufficient to stop view changes.
	//       An instance-local stable checkpoint must be created as well.

	// If the view is still the same as when the timer was set up and the segment-level checkpoint is not yet stable
	if view == state.View && !state.SegmentCheckpoint.Stable(state.Segment.Membership) {
		// Start the view change sub-protocol.
		logger.Log(logging.LevelWarn, "View change segment timer expired.", "view", state.View)
		return startViewChange(m, state, params, moduleConfig, logger)
	}

	// Do nothing otherwise.
	return nil
}

// startViewChange initiates the view change subprotocol.
// It is triggered on expiry of the SN timeout or the segment timeout.
// It constructs the PBFT view change message and creates an event requesting signing it.
func startViewChange(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	logger logging.Logger,
) error {

	// Enter the view change state and initialize a new view
	state.InViewChange = true

	var err error
	if err = state.InitView(m, params, moduleConfig, state.View+1, logger); err != nil {
		return err
	}

	// Compute the P set and Q set to be included in the ViewChange message.
	pSet, qSet := getPSetQSet(state)

	// Create a new ViewChange message.
	viewChange := pbftpbtypes.ViewChange{state.View, pSet.PbType(), qSet.PbType()} // nolint:govet
	//viewChange := pbftViewChangeMsg(state.view, pSet, qSet)

	logger.Log(logging.LevelWarn, "Starting view change.", "view", state.View)

	// Request a signature for the newly created ViewChange message.
	// Operation continues on reception of the SignResult event.
	cryptopbdsl.SignRequest(m,
		moduleConfig.Crypto,
		common.SerializeViewChangeForSigning(&viewChange),
		&viewChange,
	)

	return nil
}

// applyMsgSignedViewChange applies a signed view change message.
// The only thing it does is request verification of the signature.
func applyMsgSignedViewChange(m dsl.Module, moduleConfig common2.ModuleConfig, svc *pbftpbtypes.SignedViewChange, from t.NodeID) {
	viewChange := svc.ViewChange
	cryptopbdsl.VerifySig(
		m,
		moduleConfig.Crypto,
		common.SerializeViewChangeForSigning(viewChange),
		svc.Signature,
		from,
		svc,
	)
}

func applyVerifiedViewChange(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	svc *pbftpbtypes.SignedViewChange,
	from t.NodeID,
	logger logging.Logger,
) {
	logger.Log(logging.LevelDebug, "Received ViewChange.", "sender", from)

	// Convenience variables.
	vc := svc.ViewChange

	// Ignore message if it is from an old view.
	if vc.View < state.View {
		logger.Log(logging.LevelDebug, "Ignoring ViewChange from old view.",
			"sender", from,
			"vcView", vc.View,
			"localView", state.View,
		)
		return
	}

	// Discard ViewChange message if this node is not the primary for the referenced view
	primary := state.Segment.PrimaryNode(vc.View)
	if params.OwnID != primary {
		logger.Log(logging.LevelDebug, "Ignoring ViewChange. Not the primary of view",
			"sender", from,
			"vcView", vc.View,
			"primary", primary,
		)
		return
	}

	// Look up the state associated with the view change sub-protocol.
	vcstate := getViewChangeState(state, vc.View)

	// If enough ViewChange messages had been received already, ignore the message just received.
	if vcstate.EnoughViewChanges() {
		logger.Log(logging.LevelDebug, "Ignoring ViewChange message, have enough already", "from", from)
		return
	}

	// Update the view change state by the received ViewChange message.
	vcstate.AddSignedViewChange(svc, from, logger)

	logger.Log(logging.LevelDebug, "Added ViewChange.", "numViewChanges", len(vcstate.SignedViewChanges))

	// If enough ViewChange messages have been received
	if vcstate.EnoughViewChanges() {
		logger.Log(logging.LevelDebug, "Received enough ViewChanges.")

		// Fill in empty Preprepare messages for all sequence numbers where nothing was prepared in the old view.
		emptyPreprepareData := vcstate.SetEmptyPreprepares(vc.View, state.Segment.Proposals)

		// Request hashing of the new Preprepare messages
		hasherpbdsl.Request(m,
			moduleConfig.Hasher,
			emptyPreprepareData,
			&vc.View)
	}

	// TODO: Consider checking whether a quorum of valid view change messages has been almost received
	//       and if yes, sending a ViewChange as well if it is the last one missing.

}

func applyMsgPreprepareRequest(
	m dsl.Module,
	state *common.State,
	moduleConfig common2.ModuleConfig,
	digest []byte,
	sn tt.SeqNr,
	from t.NodeID,
) {
	if preprepare := state.LookUpPreprepare(sn, digest); preprepare != nil {

		// If the requested Preprepare message is available, send it to the originator of the request.
		// No need for periodic re-transmission.
		// In the worst case, dropping of these messages may result in another view change,
		// but will not compromise correctness.
		transportpbdsl.SendMessage(
			m,
			moduleConfig.Net,
			pbftpbmsgs.MissingPreprepare(moduleConfig.Self, preprepare),
			[]t.NodeID{from})

	}

	// If the requested Preprepare message is not available, ignore the request.
}

func applyMsgMissingPreprepare(
	m dsl.Module,
	state *common.State,
	moduleConfig common2.ModuleConfig,
	preprepare *pbftpbtypes.Preprepare,
) {

	// Ignore preprepare if received in the meantime or if view has already advanced.
	// This check is technically redundant, as it is (and must be) performed also after the Preprepare is hashed.
	// However, it might prevent some unnecessary hash computation if performed here as well.
	vcstate, view := latestPendingVCState(state)
	if pp, ok := vcstate.Preprepares[preprepare.Sn]; (ok && pp != nil) || view < state.View {
		return
	}

	// Request a hash of the received preprepare message.
	hasherpbdsl.RequestOne(
		m,
		moduleConfig.Hasher,
		common.SerializePreprepareForHashing(preprepare),
		&pbftpbtypes.MissingPreprepare{Preprepare: preprepare},
	)
}

func sendNewView(
	m dsl.Module,
	state *common.State,
	moduleConfig common2.ModuleConfig,
	view ot.ViewNr,
	vcState *common.PbftViewChangeState,
	logger logging.Logger,
) {

	logger.Log(logging.LevelDebug, "Sending NewView.")

	// Extract SignedViewChanges and their senders from the view change state.
	viewChangeSenders := make([]t.NodeID, 0, len(vcState.SignedViewChanges))
	signedViewChanges := make([]*pbftpbtypes.SignedViewChange, 0, len(vcState.SignedViewChanges))
	maputil.IterateSorted(
		vcState.SignedViewChanges,
		func(sender t.NodeID, signedViewChange *pbftpbtypes.SignedViewChange) bool {
			viewChangeSenders = append(viewChangeSenders, sender)
			signedViewChanges = append(signedViewChanges, signedViewChange)
			return true
		},
	)

	// Extract re-proposed Preprepares and their corresponding sequence numbers from the view change state.
	preprepareSeqNrs := make([]tt.SeqNr, 0, len(vcState.Preprepares))
	preprepares := make([]*pbftpbtypes.Preprepare, 0, len(vcState.Preprepares))
	maputil.IterateSorted(vcState.Preprepares, func(sn tt.SeqNr, preprepare *pbftpbtypes.Preprepare) bool {
		preprepareSeqNrs = append(preprepareSeqNrs, sn)
		preprepares = append(preprepares, preprepare)
		return true
	})

	// Construct and send the NewView message.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	transportpbdsl.SendMessage(
		m,
		moduleConfig.Net,
		pbftpbmsgs.NewView(
			moduleConfig.Self,
			view,
			t.NodeIDSlicePb(viewChangeSenders),
			signedViewChanges,
			preprepareSeqNrs,
			preprepares,
		),
		state.Segment.NodeIDs(),
	)
}

func applyMsgNewView(
	m dsl.Module,
	state *common.State,
	moduleConfig common2.ModuleConfig,
	newView *pbftpbtypes.NewView,
	from t.NodeID,
) {

	// Ignore message if the sender is not the primary of the view.
	if from != state.Segment.PrimaryNode(newView.View) {
		return
	}

	// Assemble request for checking signatures on the contained ViewChange messages.
	viewChangeData := make([]*cryptopbtypes.SignedData, len(newView.SignedViewChanges))
	signatures := make([][]byte, len(newView.SignedViewChanges))
	for i, signedViewChange := range newView.SignedViewChanges {
		viewChangeData[i] = common.SerializeViewChangeForSigning(signedViewChange.ViewChange)
		signatures[i] = signedViewChange.Signature
	}

	// Request checking of signatures on the contained ViewChange messages
	cryptopbdsl.VerifySigs(
		m,
		moduleConfig.Crypto,
		viewChangeData,
		signatures,
		t.NodeIDSlice(newView.ViewChangeSenders),
		newView,
	)

}

// ============================================================
// Auxiliary functions
// ============================================================

func validFreshPreprepare(preprepare *pbftpbtypes.Preprepare, view ot.ViewNr, sn tt.SeqNr) bool {
	return preprepare.Aborted &&
		preprepare.Sn == sn &&
		preprepare.View == view
}

// viewChangeState returns the state of the view change sub-protocol associated with the given view,
// allocating the associated data structures as needed.
func getViewChangeState(state *common.State, view ot.ViewNr) *common.PbftViewChangeState {

	if vcs, ok := state.ViewChangeStates[view]; ok {
		// If a view change state is already present, return it.
		return vcs
	}

	// If no view change state is yet associated with this view, allocate a new one and return it.
	state.ViewChangeStates[view] = common.NewPbftViewChangeState(state.Segment.SeqNrs(), state.Segment.Membership)

	return state.ViewChangeStates[view]
}

// Returns the view change state with the highest view number that received enough view change messages
// (along with the view number itself).
// If there is no view change state with enough ViewChange messages received, returns nil.
func latestPendingVCState(state *common.State) (*common.PbftViewChangeState, ot.ViewNr) {

	// View change state with the highest view that received enough ViewChange messages and its view number.
	var vcstate *common.PbftViewChangeState
	var view ot.ViewNr

	// Find and return the view change state with the highest view number that received enough ViewChange messages.
	for v, s := range state.ViewChangeStates {
		if s.EnoughViewChanges() && (state == nil || v > view) {
			vcstate, view = s, v
		}
	}
	return vcstate, view
}

// ============================================================
// ViewChange message construction
// ============================================================

// getPSetQSet computes the P set and Q set for the construction of a PBFT view change message.
// Note that this representation of the PSet and QSet is internal to the protocol implementation
// and cannot be directly used in a view change message.
// They must first be transformed to a serializable representation that adheres to the message format.
func getPSetQSet(state *common.State) (pSet common.ViewChangePSet, qSet common.ViewChangeQSet) {
	// Initialize the PSet.
	pSet = make(map[tt.SeqNr]*pbftpbtypes.PSetEntry)

	// Initialize the QSet.
	qSet = make(map[tt.SeqNr]map[string]ot.ViewNr)

	// For each sequence number, compute the PSet and the QSet.
	for _, sn := range state.Segment.SeqNrs() {

		// Initialize QSet.
		// (No need to initialize the PSet, as, unlike the PSet,
		// the QSet may hold multiple values for the same sequence number.)
		qSet[sn] = make(map[string]ot.ViewNr)

		// Traverse all previous views.
		// The direction of iteration is important, so the values from newer views can overwrite values from older ones.
		for view := ot.ViewNr(0); view < state.View; view++ {
			// Skip views that the node did not even enter
			if slots, ok := state.Slots[view]; ok {

				// Get the PbftSlot of sn in view (convenience variable)
				slot := slots[sn]

				// If a value was prepared for sn in view, add the corresponding entry to the PSet.
				// If there was an entry corresponding to an older view, it will be overwritten.
				if slot.Prepared {
					pSet[sn] = &pbftpbtypes.PSetEntry{
						Sn:     sn,
						View:   view,
						Digest: slot.Digest,
					}
				}

				// If a value was preprepared for sn in view, add the corresponding entry to the QSet.
				// If the same value has been preprepared in an older view, its entry will be overwritten.
				if slot.Preprepared {
					qSet[sn][string(slot.Digest)] = view
				}
			}
		}
	}
	return pSet, qSet
}

// askForMissingPreprepares requests the Preprepare messages that are part of a new view.
// The new primary might have received a prepare certificate from other nodes in the ViewChange messages they sent
// and thus the new primary has to re-propose the corresponding availability certificate
// by including the corresponding Preprepare message in the NewView message.
// However, the new primary might not have all the corresponding Preprepare messages,
// in which case it calls this function.
// Note that the requests for missing Preprepare messages need not necessarily be periodically re-transmitted.
// If they are dropped, the new primary will simply never send a NewView message
// and will be succeeded by another primary after another view change.
func askForMissingPreprepares(m dsl.Module, moduleConfig common2.ModuleConfig, vcState *common.PbftViewChangeState) {
	for sn, digest := range vcState.Reproposals {
		if len(digest) > 0 && vcState.Preprepares[sn] == nil {
			transportpbdsl.SendMessage(
				m,
				moduleConfig.Net,
				pbftpbmsgs.PreprepareRequest(moduleConfig.Self, digest, sn),
				vcState.PrepreparedIDs[sn],
			) // TODO be smarter about this eventually, not asking everyone at once.
		}
	}
}

func getMsgView(msgPb proto.Message) (ot.ViewNr, error) {
	switch msg := msgPb.(type) {
	case *pbftpb.Preprepare:
		return ot.ViewNr(msg.View), nil
	case *pbftpb.Prepare:
		return ot.ViewNr(msg.View), nil
	case *pbftpb.Commit:
		return ot.ViewNr(msg.View), nil
	default:
		return 0, es.Errorf("invalid PBFT message for view extraction: %T (%v)", msg, msg)
	}
}
