package goodcase

import (
	"fmt"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"
	es "github.com/go-errors/errors"
	"reflect"

	common2 "github.com/filecoin-project/mir/pkg/orderers/common"
	"github.com/filecoin-project/mir/pkg/orderers/internal/parts/catchup"
	availabilitypbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	isspbdsl "github.com/filecoin-project/mir/pkg/pb/isspb/dsl"
	pbftpbdsl "github.com/filecoin-project/mir/pkg/pb/pbftpb/dsl"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"

	ot "github.com/filecoin-project/mir/pkg/orderers/types"

	"github.com/filecoin-project/mir/pkg/orderers/internal/common"

	"github.com/filecoin-project/mir/pkg/dsl"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	ppvpbdsl "github.com/filecoin-project/mir/pkg/pb/ordererpb/pprepvalidatorpb/dsl"
	pbftpbevents "github.com/filecoin-project/mir/pkg/pb/pbftpb/events"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/logging"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

func IncludeGoodCase(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	logger logging.Logger,
) {

	// UponInit take all the actions resulting from the PBFT State's initial state.
	// The Init event is expected to be the first event applied to the State,
	// except for events read from the WAL at startup, which are expected to be applied even before the Init event.
	// (At this time, the WAL is not used. TODO: Update this when wal is implemented.)
	eventpbdsl.UponInit(m, func() error {
		// Initialize the first PBFT view
		if err := state.InitView(m, params, moduleConfig, 0, logger); err != nil {
			return err
		}

		eventpbdsl.TimerDelay(
			m,
			moduleConfig.Timer,
			[]*eventpbtypes.Event{pbftpbevents.ProposeTimeout(moduleConfig.Self, 1)},
			types.Duration(params.Config.MaxProposeDelay),
		)

		// Set up timer for the first proposal.
		return nil
	})

	// UponNewCert process the new availability certificate ready to be proposed.
	// This event is triggered by the availability module in response to the CertRequest event produced by this State.
	apbdsl.UponNewCert(m, func(cert *availabilitypbtypes.Cert, _ *struct{}) error {
		// Clear flag that was set in requestNewCert(), so that new certificates can be requested if necessary.
		state.Proposal.CertRequested = false

		if state.Proposal.CertRequestedView == state.View {
			// If the protocol is still in the same PBFT view as when the certificate was requested,
			// propose the received certificate.

			certBytes, err := proto.Marshal(cert.Pb())
			if err != nil {
				return es.Errorf("error marshalling certificate: %w", err)
			}

			err = propose(m, state, params, moduleConfig, certBytes, logger)
			if err != nil {
				return es.Errorf("failed to propose: %w", err)
			}
		} else { // nolint:staticcheck,revive
			// If the PBFT view advanced since the certificate was requested,
			// do not propose the certificate and resurrect the transactions it contains.
			// eventsOut.PushBack(state.eventService.SBEvent(SBResurrectBatchEvent(batch.Batch)))
			// TODO: Implement resurrection!
		}

		return nil
	})

	// UponResultOne process the preprepare message and send a Prepares message to all nodes.
	hasherpbdsl.UponResultOne(m, func(digest []byte, context *pbftpbtypes.Preprepare) error {

		if params.AccountabilityMode == common.AllAccountable {

			data := serializePrepareForSigning(&pbftpbtypes.Prepare{
				Sn:     context.Sn,
				View:   state.View,
				Digest: digest,
			})

			//sign prepare
			cryptopbdsl.SignRequest(m,
				moduleConfig.Crypto,
				&cryptopbtypes.SignedData{data},
				&signPrepareContext{
					preprepare: context,
					digest:     digest,
				},
			)
		} else {
			// Send an unsigned Prepares message.
			return sendPrepare(m,
				state,
				params,
				moduleConfig,
				digest,
				context,
				nil,
				logger)
		}

		return nil
	})

	hasherpbdsl.UponResultOne(m, func(digest []byte, context *signPreprepareContext) error {
		cryptopbdsl.SignRequest(m,
			moduleConfig.Crypto,
			&cryptopbtypes.SignedData{serializePreprepareForSigning(digest, context.preprepare)},
			context)

		return nil
	})

	cryptopbdsl.UponSignResult(m, func(signature []byte, context *signPrepareContext) error {
		// Send a signed Prepares message.
		return sendPrepare(m,
			state,
			params,
			moduleConfig,
			context.digest,
			context.preprepare,
			signature,
			logger)
	})

	cryptopbdsl.UponSignResult(m, func(signature []byte, context *signPreprepareContext) error {
		// Send a Prepares message.
		// Send signature on digest but not digest itself (it would be redundant since receives need to calculate digest anyway)
		// Signature on digest is required for 'easy' verification when leader's signature attached to commit messages
		sendPreprepare(m, state, params, moduleConfig, context.preprepare, logger)
		return nil
	})

	hasherpbdsl.UponResultOne(m, func(digest []byte, context *validatePreprepareContext) error {
		cryptopbdsl.VerifySig(m,
			moduleConfig.Crypto,
			&cryptopbtypes.SignedData{
				serializePreprepareForSigning(digest, context.preprepare),
			},
			context.preprepare.Signature,
			context.from,
			context)
		return nil
	})

	cryptopbdsl.UponSigVerified(m, func(from t.NodeID, err error, context *validatePreprepareContext) error {
		if err != nil {
			logger.Log(logging.LevelWarn, "preprepare signature from %v is invalid", from)
			return nil
		}

		ppvpbdsl.Validatepreprepare(m,
			moduleConfig.PPrepValidator,
			context.preprepare,
			context)
		return nil
	})

	cryptopbdsl.UponSigVerified(m, func(from t.NodeID, err error, context *validatePrepareSigContext) error {
		if err != nil {
			logger.Log(logging.LevelWarn, "prepare signature from %v is invalid", from)
			return nil
		}

		applyMsgPrepareSigValidated(
			m,
			state,
			params,
			moduleConfig,
			context.prepare,
			from,
			logger)

		return nil
	})

	pbftpbdsl.UponProposeTimeout(m, func(proposeTimeout uint64) error {
		return applyProposeTimeout(m, state, params, moduleConfig, int(proposeTimeout), logger)
	})

	pbftpbdsl.UponPreprepareReceived(m, func(from t.NodeID, sn tt.SeqNr, view ot.ViewNr, data []byte, aborted bool, _ []byte) error {
		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		preprepare := &pbftpbtypes.Preprepare{
			Sn:      sn,
			View:    view,
			Data:    data,
			Aborted: aborted,
		}
		ApplyMsgPreprepare(m, params, moduleConfig, preprepare, from, logger)
		return nil
	})

	ppvpbdsl.UponPreprepareValidated(m, func(err error, c *validatePreprepareContext) error {
		ApplyMsgPreprepareValidated(m, state, moduleConfig, c.preprepare, c.from, logger, err)
		return nil
	})

	pbftpbdsl.UponPrepareReceived(m, func(from t.NodeID, sn tt.SeqNr, view ot.ViewNr, digest []byte, _ []byte) error {
		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}

		prepare := &pbftpbtypes.Prepare{
			Sn:     sn,
			View:   view,
			Digest: digest,
		}
		applyMsgPrepare(m, state, params, moduleConfig, prepare, from, logger)
		return nil
	})

	pbftpbdsl.UponCommitReceived(m, func(from t.NodeID, sn tt.SeqNr, view ot.ViewNr, digest []byte, preprepare *pbftpbtypes.Preprepare) error {
		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		commit := &pbftpbtypes.Commit{
			Sn:         sn,
			View:       view,
			Digest:     digest,
			Preprepare: preprepare,
		}
		applyMsgCommit(m, state, params, moduleConfig, commit, from, logger)
		return nil
	})

	// UponResultOne process the preprepare message and send a Prepares message to all nodes.
	hasherpbdsl.UponResultOne(m, func(digest []byte, context *validateCommitContext) error {
		if !reflect.DeepEqual(digest, context.commit.Digest) {
			logger.Log(logging.LevelWarn, "Ignoring Commit message with different digest than attached Prepare.",
				"sn", context.commit.Sn, "from", context.from)
			return nil
		}

		context.digest = digest
		data := serializePreprepareForSigning(digest, context.commit.Preprepare)
		//sign prepare
		cryptopbdsl.VerifySig(m,
			moduleConfig.Crypto,
			&cryptopbtypes.SignedData{data},
			context.commit.Preprepare.Signature,
			context.primary, // leader
			context,
		)
		return nil

	})

	cryptopbdsl.UponSigVerified(m, func(primary t.NodeID, err error, context *validateCommitContext) error {
		if err != nil {
			logger.Log(logging.LevelWarn, "Preprepare signature attached to commit from %v is invalid", context.from)
			return nil
		}

		commit := context.commit // convenience variable

		// Preprocess message, looking up the corresponding pbftSlot.
		slot := preprocessMessage(state, commit.Sn, commit.View, commit.Pb(), context.from, logger)
		if slot == nil {
			// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
			return nil
		}

		if slot.PreprepareForAcc != nil {
			if !reflect.DeepEqual(slot.Preprepare.Signature, commit.Preprepare.Signature) {
				logger.Log(logging.LevelWarn, "Ignoring Commit message with different signature than attached Preprepare.",
					"sn", commit.Sn, "from", context.from)
				//TODO ACC-UPDATE POM FOUND!
				// Here is where we should inform of equivocations and send to others stored prepare messages so that they find all equivocations and report them
				// 3 new events/messages should be implemented here: Notify of equivocation:
				// 1- Notify of PoM to application layer?,
				// 2- Send PoM for Leader?,
				// 3- send stored prepare messages (this one only if AllAccountable),

				return nil
			}
		}

		slot.PreprepareForAcc = commit.Preprepare // first one slips through because we cannot compare with anything, but all following commits will have to verify against this
		// (which is enough to guarantee accountability unless weightOf(from)>=strongQuorum()
		applyMsgCommitSigValidated(m, state, params, moduleConfig, commit, context.from, logger)

		return nil
	})

}

func sendPrepare(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	digest []byte,
	preprepare *pbftpbtypes.Preprepare,
	signature []byte,
	logger logging.Logger,
) error {

	// Stop processing the Preprepare if view advanced in the meantime.
	if preprepare.View < state.View {
		return nil
	}

	// Save the digest of the Preprepare message and mark the slot as preprepared.
	slot := state.Slots[state.View][preprepare.Sn]
	slot.Digest = digest
	slot.Preprepared = true
	// Send a Prepares message.
	transportpbdsl.SendMessage(
		m,
		moduleConfig.Net,
		pbftpbmsgs.Prepare(moduleConfig.Self, preprepare.Sn, state.View, digest, signature),
		state.Segment.NodeIDs(),
	)
	// Advance the state of the PbftSlot even more if necessary
	// (potentially sending a Commit message or even delivering).
	// This is required for the case when the Preprepare message arrives late.
	advanceSlotState(m, state, params, moduleConfig, slot, preprepare.Sn, logger)
	return nil
}

// applyProposeTimeout applies the event of the proposal timeout firing.
// It updates the proposal state accordingly and triggers a new proposal if possible.
func applyProposeTimeout(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	numProposals int,
	logger logging.Logger,
) error {

	// If we are still waiting for this timeout
	if numProposals > state.Proposal.ProposalTimeout {

		// Save the number of proposals for which the timeout fired.
		state.Proposal.ProposalTimeout = numProposals

		// If this was the last bit missing, start a new proposal.
		if canPropose(state, params) {

			// If a proposal already has been set as a parameter of the segment, propose it directly.
			sn := state.Segment.SeqNrs()[state.Proposal.ProposalsMade]
			if proposal := state.Segment.Proposals[sn]; proposal != nil {
				return propose(m, state, params, moduleConfig, proposal, logger)
			}

			// Otherwise, obtain a fresh availability certificate first.
			requestNewCert(m, state, moduleConfig)
		}
	}

	return nil
}

// requestNewCert asks (by means of a CertRequest event) the availability module to provide a new availability certificate.
// When the certificate is ready, it must be passed to the State using the CertReady event.
func requestNewCert(m dsl.Module, state *common.State, moduleConfig common2.ModuleConfig) {

	// Set a flag indicating that a certificate has been requested,
	// so that no new certificates will be requested before the reception of this one.
	// It will be cleared when CertReady is received.
	state.Proposal.CertRequested = true

	// Remember the view in which the certificate has been requested
	// to make sure that we are still in the same view when the certificate becomes ready.
	state.Proposal.CertRequestedView = state.View

	// Emit the CertRequest event.
	// Operation continues on reception of the CertReady event.
	apbdsl.RequestCert(m, moduleConfig.Ava, &struct{}{})
}

// propose proposes a new availability certificate by sending a Preprepare message.
// propose assumes that the state of the PBFT State allows sending a new proposal
// and does not perform any checks in this regard.
func propose(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	data []byte,
	logger logging.Logger,
) error {

	// Send a Preprepare message.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	sn := state.Segment.SeqNrs()[state.Proposal.ProposalsMade]
	preprepare := &pbftpbtypes.Preprepare{
		sn,
		state.View,
		data,
		false,
		nil,
	}

	if params.AccountabilityMode == common.AllAccountable ||
		params.AccountabilityMode == common.LeaderAccountable {
		//first calculate digest, we want to sign the digest for verification when nodes attach the leader's preprepare to their commit
		hasherpbdsl.RequestOne(
			m,
			moduleConfig.Hasher,
			common.SerializePreprepareForHashing(preprepare),
			&signPreprepareContext{
				preprepare: preprepare,
			})
	} else {
		sendPreprepare(m, state, params, moduleConfig, preprepare, logger)
	}

	return nil
}

func sendPreprepare(m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	preprepare *pbftpbtypes.Preprepare,
	logger logging.Logger) {

	// Update proposal counter.
	state.Proposal.ProposalsMade++

	// Log debug message.
	logger.Log(logging.LevelDebug, "Proposing.", "sn", preprepare.Sn)

	transportpbdsl.SendMessage(m,
		moduleConfig.Net,
		pbftpbmsgs.Preprepare(moduleConfig.Self, preprepare.Sn, preprepare.View, preprepare.Data, preprepare.Aborted, preprepare.Signature),
		state.Segment.NodeIDs())

	// Set up a new timer for the next proposal.
	timeoutEvent := pbftpbevents.ProposeTimeout(
		moduleConfig.Self,
		uint64(state.Proposal.ProposalsMade+1))

	eventpbdsl.TimerDelay(m,
		moduleConfig.Timer,
		[]*eventpbtypes.Event{timeoutEvent},
		types.Duration(params.Config.MaxProposeDelay),
	)
}

// ApplyMsgPreprepare applies a received preprepare message.
// It performs the necessary checks and, if successful, submits it for hashing.
func ApplyMsgPreprepare(
	m dsl.Module,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	preprepare *pbftpbtypes.Preprepare,
	from t.NodeID,
	logger logging.Logger,
) {

	if params.AccountabilityMode == common.LeaderAccountable ||
		params.AccountabilityMode == common.AllAccountable {
		hasherpbdsl.RequestOne(
			m,
			moduleConfig.Hasher,
			common.SerializePreprepareForHashing(preprepare),
			&validatePreprepareContext{
				preprepare: preprepare,
				from:       from,
			},
		)
		return

	} else if preprepare.Signature != nil {
		logger.Log(logging.LevelWarn, "Ignoring Preprepare message with signature while no signature expected")
		return
	}
	ppvpbdsl.Validatepreprepare(m, moduleConfig.PPrepValidator, preprepare, &validatePreprepareContext{
		preprepare: preprepare,
		from:       from,
	})
}

func ApplyMsgPreprepareValidated(
	m dsl.Module,
	state *common.State,
	moduleConfig common2.ModuleConfig,
	preprepare *pbftpbtypes.Preprepare,
	from t.NodeID,
	logger logging.Logger,
	err error,
) {

	// Convenience variable
	sn := preprepare.Sn

	if err != nil {
		logger.Log(logging.LevelWarn, "Ignoring Preprepare message with invalid proposal.",
			"sn", sn, "from", from, "err", err)
		return
	}

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := preprocessMessage(state, sn, preprepare.View, preprepare.Pb(), from, logger)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return
	}

	// Check that this is the first Preprepare message received.
	// Note that checking the orderer.Preprepared flag instead would be incorrect,
	// as that flag is only set after computing the Preprepare message's digest.
	if slot.Preprepare != nil {
		logger.Log(logging.LevelDebug, "Ignoring Preprepare message. Already preprepared or prepreparing.",
			"sn", sn, "from", from)
		return
	}

	// Check whether the sender is the legitimate leader of the segment in the current view.
	primary := state.Segment.PrimaryNode(state.View)
	if from != primary {
		logger.Log(logging.LevelWarn, "Ignoring Preprepare message. Invalid Leader.",
			"expectedLeader", primary,
			"sender", from,
		)
		return
	}

	slot.Preprepare = preprepare
	slot.PreprepareForAcc = preprepare

	// Request the computation of the hash of the Preprepare message.
	//TODO This can be optimized in the case of at least leader accountable because
	// hash already calculated
	hasherpbdsl.RequestOne(
		m,
		moduleConfig.Hasher,
		common.SerializePreprepareForHashing(preprepare),
		preprepare,
	)
}

// applyMsgPrepare applies a received preprepare message.
// It performs the necessary checks and, if successful,
// may trigger additional events like the sending of a Commit message.
func applyMsgPrepare(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	prepare *pbftpbtypes.Prepare,
	from t.NodeID,
	logger logging.Logger,
) {

	if params.AccountabilityMode == common.AllAccountable {
		cryptopbdsl.VerifySig(m,
			moduleConfig.Crypto,
			&cryptopbtypes.SignedData{serializePrepareForSigning(prepare)},
			prepare.Signature,
			from,
			&validatePrepareSigContext{
				prepare: prepare,
			})
		return
	} else if prepare.Signature != nil {
		logger.Log(logging.LevelWarn, "Ignoring Prepares message with signature while no signature expected")
		return
	}

	applyMsgPrepareSigValidated(
		m,
		state,
		params,
		moduleConfig,
		prepare,
		from,
		logger)

}

func applyMsgPrepareSigValidated(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	prepare *pbftpbtypes.Prepare,
	from t.NodeID,
	logger logging.Logger,
) {
	if prepare.Digest == nil {
		logger.Log(logging.LevelWarn, "Ignoring Prepares message with nil digest.")
		return
	}

	// Convenience variable
	sn := prepare.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := preprocessMessage(state, sn, prepare.View, prepare.Pb(), from, logger)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return
	}

	// Check if a Prepares message has already been received from this node in the current view.
	if _, ok := slot.Prepares[from]; ok {
		logger.Log(logging.LevelDebug, "Ignoring Prepares message. Already received in this view.",
			"sn", sn, "from", from, "view", state.View)
		return
	}

	// Save the received Prepares message and advance the slot state
	// (potentially sending a Commit message or even delivering).
	slot.Prepares[from] = prepare
	advanceSlotState(m, state, params, moduleConfig, slot, sn, logger)
}

// applyMsgCommit applies a received commit message.
// It performs the necessary checks and, if successful,
// may trigger additional events like delivering the corresponding certificate.
func applyMsgCommit(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	commit *pbftpbtypes.Commit,
	from t.NodeID,
	logger logging.Logger,
) {
	if commit.Digest == nil {
		logger.Log(logging.LevelWarn, "Ignoring Commit message with nil digest.")
		return
	}

	if params.AccountabilityMode == common.AllAccountable ||
		params.AccountabilityMode == common.LeaderAccountable {
		if commit.Preprepare == nil {
			logger.Log(logging.LevelWarn, "Ignoring Commit message with nil attached Preprepare.",
				"sn", commit.Sn, "from", from)
			return
		}
		if commit.Preprepare.Sn != commit.Sn {
			logger.Log(logging.LevelWarn, "Ignoring Commit message with different sequence number than attached Prepare.",
				"sn", commit.Sn, "from", from)
			return
		}

		if commit.Preprepare.View != commit.View {
			logger.Log(logging.LevelWarn, "Ignoring Commit message with different view than attached Prepare.",
				"sn", commit.Sn, "from", from)
			return
		}

		hasherpbdsl.RequestOne(
			m,
			moduleConfig.Hasher,
			common.SerializePreprepareForHashing(commit.Preprepare),
			&validateCommitContext{
				commit:  commit,
				from:    from,
				primary: state.Segment.PrimaryNode(commit.View),
			})

	} else if commit.Preprepare != nil {
		logger.Log(logging.LevelWarn, "Ignoring Commit message with attached Prepare while no Prepare expected")
		return
	}

	applyMsgCommitSigValidated(m,
		state,
		params,
		moduleConfig,
		commit,
		from,
		logger)
}
func applyMsgCommitSigValidated(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	commit *pbftpbtypes.Commit,
	from t.NodeID,
	logger logging.Logger,
) {

	// Convenience variable
	sn := commit.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := preprocessMessage(state, sn, commit.View, commit.Pb(), from, logger)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return
	}

	// Check if a Commit message has already been received from this node in the current view.
	if _, ok := slot.CommitDigests[from]; ok {
		logger.Log(logging.LevelDebug, "Ignoring Commit message. Already received in this view.",
			"sn", sn, "from", from, "view", state.View)
		return
	}

	// Save the received Commit message and advance the slot state
	// (potentially delivering the corresponding certificate and its successors).
	slot.CommitDigests[from] = commit.Digest
	advanceSlotState(m, state, params, moduleConfig, slot, sn, logger)
}

func ApplyBufferedMsg(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	msgPb proto.Message,
	from t.NodeID,
	logger logging.Logger,
) {
	switch msg := msgPb.(type) {
	case *pbftpb.Preprepare:
		ApplyMsgPreprepare(m, params, moduleConfig, pbftpbtypes.PreprepareFromPb(msg), from, logger)
	case *pbftpb.Prepare:
		applyMsgPrepare(m, state, params, moduleConfig, pbftpbtypes.PrepareFromPb(msg), from, logger)
	case *pbftpb.Commit:
		applyMsgCommit(m, state, params, moduleConfig, pbftpbtypes.CommitFromPb(msg), from, logger)
	}
}

// preprocessMessage performs basic checks on a PBFT protocol message based the associated view and sequence number.
// If the message is invalid or outdated, preprocessMessage simply returns nil.
// If the message is from a future view, preprocessMessage will try to buffer it for later processing and returns nil.
// If the message can be processed, preprocessMessage returns the pbftSlot tracking the corresponding state.
// The slot returned by preprocessMessage always belongs to the current view.
func preprocessMessage(
	state *common.State,
	sn tt.SeqNr,
	view ot.ViewNr,
	msg proto.Message,
	from t.NodeID,
	logger logging.Logger,
) *common.PbftSlot {

	if view < state.View {
		// Ignore messages from old views.
		logger.Log(logging.LevelDebug, "Ignoring message from old view.",
			"sn", sn, "from", from, "msgView", view, "localView", state.View, "type", fmt.Sprintf("%T", msg))
		return nil
	} else if view > state.View || state.InViewChange {
		// If message is from a future view, buffer it.
		logger.Log(logging.LevelDebug, "Buffering message.",
			"sn", sn, "from", from, "msgView", view, "localView", state.View, "type", fmt.Sprintf("%T", msg))
		state.MessageBuffers[from].Store(msg)
		return nil
		// TODO: When view change is implemented, get the messages out of the buffer.
	} else if slot, ok := state.Slots[state.View][sn]; !ok {
		// If the message is from the current view, look up the slot concerned by the message
		// and check if the sequence number is assigned to this PBFT instance.
		logger.Log(logging.LevelDebug, "Ignoring message. Wrong sequence number.",
			"sn", sn, "from", from, "msgView", view)
		return nil
	} else if slot.Committed {
		// Check if the slot has already been committed (this can happen with state transfer).
		return nil
	} else {
		return slot
	}
}

// canPropose returns true if the current state of the PBFT State
// allows for a new availability certificate to be proposed.
// Note that "new availability certificate" means a "fresh" certificate
// proposed during normal operation outside of view change.
// Proposals part of a new view message during a view change do not call this function and are treated separately.
func canPropose(state *common.State, params *common.ModuleParams) bool {
	return params.OwnID == state.Segment.Leader && // Only the leader can propose

		// No regular proposals can be made after a view change.
		// This is specific for the SB-version of PBFT used in ISS and deviates from the standard PBFT protocol.
		state.View == 0 &&

		// A new certificate must not have been requested (if it has, we are already in the process of proposing).
		!state.Proposal.CertRequested &&

		// There must still be a free sequence number for which a proposal can be made.
		state.Proposal.ProposalsMade < state.Segment.Len() &&

		// The proposal timeout must have passed.
		state.Proposal.ProposalTimeout > state.Proposal.ProposalsMade &&

		// No proposals can be made while in view change.
		!state.InViewChange
}

// advanceSlotState checks whether the state of the PbftSlot can be advanced.
// If it can, advanceSlotState updates the state of the PbftSlot and returns a list of Events that result from it.
// Requires the PBFT instance as an argument to use it to generate the proper events.
func advanceSlotState(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	slot *common.PbftSlot,
	sn tt.SeqNr,
	logger logging.Logger,
) {
	// If the slot just became prepared, send the Commit message.
	if !slot.Prepared && slot.CheckPrepared() {
		slot.Prepared = true

		var preprepare *pbftpbtypes.Preprepare
		if params.AccountabilityMode == common.AllAccountable ||
			params.AccountabilityMode == common.LeaderAccountable {
			preprepare = slot.Preprepare
		}

		transportpbdsl.SendMessage(
			m,
			moduleConfig.Net,
			pbftpbmsgs.Commit(moduleConfig.Self, sn, state.View, slot.Digest, preprepare),
			state.Segment.NodeIDs())
	}

	// If the slot just became committed, reset SN timeout and deliver the certificate.
	if !slot.Committed && slot.CheckCommitted() {

		// Mark slot as committed.
		slot.Committed = true

		// Set a new SN timeout (unless the segment is finished with a stable checkpoint).
		// Note that we set a new SN timeout even if everything has already been committed.
		// In such a case, we expect a quorum of other nodes to also commit everything and send a Done message
		// before the timeout fires.
		// The timeout event contains the current view and the number of committed slots.
		// It will be ignored if any of those values change by the time the timer fires
		// or if a quorum of nodes confirms having committed all certificates.
		if !state.SegmentCheckpoint.Stable(state.Segment.Membership) {
			eventpbdsl.TimerDelay(
				m,
				moduleConfig.Timer,
				[]*eventpbtypes.Event{pbftpbevents.ViewChangeSNTimeout(
					moduleConfig.Self,
					state.View,
					uint64(state.NumCommitted(state.View)))},
				types.Duration(params.Config.ViewChangeSNTimeout))
		}

		// If all certificates have been committed (i.e. this is the last certificate to commit),
		// send a Done message to all other nodes.
		// This is required for liveness, see comments for pbftSegmentChkp.
		if state.AllCommitted() {
			state.SegmentCheckpoint.SetDone()
			catchup.SendDoneMessages(m, state, params, moduleConfig, logger)
		}

		// Deliver availability certificate (will be verified by ISS)
		isspbdsl.SBDeliver(
			m,
			moduleConfig.Ord,
			sn,
			slot.Preprepare.Data,
			slot.Preprepare.Aborted,
			state.Segment.Leader,
			moduleConfig.Self,
		)
	}
}

func serializePreprepareForSigning(digest []byte, preprepare *pbftpbtypes.Preprepare) [][]byte {
	result := make([][]byte, 0, 2)
	result[0] = digest
	result[1] = preprepare.View.Bytes() //this is not being used for hashing , which is why I add it here
	return result
}

func serializePrepareForSigning(prepare *pbftpbtypes.Prepare) [][]byte {
	result := make([][]byte, 0, 3)
	result[0] = prepare.Digest
	result[1] = prepare.View.Bytes()
	result[2] = prepare.Sn.Bytes()
	return result
}

type validatePreprepareContext struct {
	preprepare *pbftpbtypes.Preprepare
	from       t.NodeID
}

type signPrepareContext struct {
	preprepare *pbftpbtypes.Preprepare
	digest     []byte
}

type signPreprepareContext struct {
	preprepare *pbftpbtypes.Preprepare
}

type validatePrepareSigContext struct {
	prepare *pbftpbtypes.Prepare
}

type validateCommitContext struct {
	commit  *pbftpbtypes.Commit
	from    t.NodeID
	digest  []byte
	primary t.NodeID
}
