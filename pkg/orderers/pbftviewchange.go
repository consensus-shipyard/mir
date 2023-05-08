package orderers

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/filecoin-project/mir/pkg/dsl"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"

	"github.com/filecoin-project/mir/pkg/iss/config"
	ot "github.com/filecoin-project/mir/pkg/orderers/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	timertypes "github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"

	"github.com/filecoin-project/mir/pkg/logging"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// Creates a new Preprepare message identical to the one given as argument,
// except for the view number being set to view.
// The Preprepare message produced by this function has the same digest as the original preprepare,
// since the view number is not used for hash computation.
func copyPreprepareToNewView(preprepare *pbftpbtypes.Preprepare, view ot.ViewNr) *pbftpbtypes.Preprepare {
	return &pbftpbtypes.Preprepare{preprepare.Sn, view, preprepare.Data, preprepare.Aborted} // nolint:govet
}

// ============================================================
// OrdererEvent handling
// ============================================================

// applyViewChangeSNTimeout applies the view change SN timeout event
// triggered some time after a value is committed.
// If nothing has been committed since, triggers a view change.
func (orderer *Orderer) applyViewChangeSNTimeout(m dsl.Module, view ot.ViewNr, numCommitted uint64) error {

	// If the view is still the same as when the timer was set up,
	// if nothing has been committed since then, and if the segment-level checkpoint is not yet stable
	if view == orderer.view &&
		int(numCommitted) == orderer.numCommitted(orderer.view) &&
		!orderer.segmentCheckpoint.Stable(len(orderer.segment.Membership.Nodes)) {

		// Start the view change sub-protocol.
		orderer.logger.Log(logging.LevelWarn, "View change SN timer expired.",
			"view", orderer.view, "numCommitted", numCommitted)
		return orderer.startViewChange(m)
	}

	// Do nothing otherwise.
	return nil
}

// applyViewChangeSegmentTimeout applies the view change segment timeout event
// triggered some time after a segment is initialized.
// If not all slots have been committed, and the view has not advanced, triggers a view change.
func (orderer *Orderer) applyViewChangeSegmentTimeout(m dsl.Module, view ot.ViewNr) error {

	// TODO: All slots being committed is not sufficient to stop view changes.
	//       An instance-local stable checkpoint must be created as well.

	// If the view is still the same as when the timer was set up and the segment-level checkpoint is not yet stable
	if view == orderer.view && !orderer.segmentCheckpoint.Stable(len(orderer.segment.Membership.Nodes)) {
		// Start the view change sub-protocol.
		orderer.logger.Log(logging.LevelWarn, "View change segment timer expired.", "view", orderer.view)
		return orderer.startViewChange(m)
	}

	// Do nothing otherwise.
	return nil
}

// startViewChange initiates the view change subprotocol.
// It is triggered on expiry of the SN timeout or the segment timeout.
// It constructs the PBFT view change message and creates an event requesting signing it.
func (orderer *Orderer) startViewChange(m dsl.Module) error {

	// Enter the view change state and initialize a new view
	orderer.inViewChange = true

	var err error
	if err = orderer.initView(m, orderer.view+1); err != nil {
		return err
	}

	// Compute the P set and Q set to be included in the ViewChange message.
	pSet, qSet := orderer.getPSetQSet()

	// Create a new ViewChange message.
	viewChange := pbftpbtypes.ViewChange{orderer.view, pSet.PbType(), qSet.PbType()} // nolint:govet
	//viewChange := pbftViewChangeMsg(orderer.view, pSet, qSet)

	orderer.logger.Log(logging.LevelWarn, "Starting view change.", "view", orderer.view)

	// Request a signature for the newly created ViewChange message.
	// Operation continues on reception of the SignResult event.
	cryptopbdsl.SignRequest(m,
		orderer.moduleConfig.Crypto,
		serializeViewChangeForSigning(&viewChange),
		&viewChange,
	)

	return nil
}

// applyViewChangeSignResult processes a newly generated signature of a ViewChange message.
// It creates a SignedViewChange message and sends it to the new leader (PBFT primary)
func (orderer *Orderer) applyViewChangeSignResult(m dsl.Module, signature []byte, viewChange *pbftpbtypes.ViewChange) {

	// Compute the primary of this view (using round-robin on the membership)
	primary := orderer.segment.PrimaryNode(viewChange.View)

	// Repeatedly send the ViewChange message. Repeat until this instance of PBFT is garbage-collected,
	// i.e., from the point of view of the PBFT protocol, effectively forever.

	eventpbdsl.TimerRepeat(m, orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{transportpbevents.SendMessage(
			orderer.moduleConfig.Net,
			pbftpbmsgs.SignedViewChange(orderer.moduleConfig.Self, viewChange, signature),
			[]t.NodeID{primary},
		)},
		timertypes.Duration(orderer.config.ViewChangeResendPeriod),
		tt.RetentionIndex(orderer.config.epochNr),
	)

}

// applyMsgSignedViewChange applies a signed view change message.
// The only thing it does is request verification of the signature.
func (orderer *Orderer) applyMsgSignedViewChange(m dsl.Module, svc *pbftpbtypes.SignedViewChange, from t.NodeID) {
	// TODO Update to use VerifySig and not VerifySigs (and corresponding Upon closure)
	viewChange := svc.ViewChange
	cryptopbdsl.VerifySig(
		m,
		orderer.moduleConfig.Crypto,
		serializeViewChangeForSigning(viewChange),
		svc.Signature,
		from,
		svc,
	)
}

func (orderer *Orderer) applyVerifiedViewChange(m dsl.Module, svc *pbftpbtypes.SignedViewChange, from t.NodeID) {
	orderer.logger.Log(logging.LevelDebug, "Received ViewChange.", "sender", from)

	// Convenience variables.
	vc := svc.ViewChange

	// Ignore message if it is from an old view.
	if vc.View < orderer.view {
		orderer.logger.Log(logging.LevelDebug, "Ignoring ViewChange from old view.",
			"sender", from,
			"vcView", vc.View,
			"localView", orderer.view,
		)
		return
	}

	// Discard ViewChange message if this node is not the primary for the referenced view
	primary := orderer.segment.PrimaryNode(vc.View)
	if orderer.ownID != primary {
		orderer.logger.Log(logging.LevelDebug, "Ignoring ViewChange. Not the primary of view",
			"sender", from,
			"vcView", vc.View,
			"primary", primary,
		)
		return
	}

	// Look up the state associated with the view change sub-protocol.
	state := orderer.getViewChangeState(vc.View)

	// If enough ViewChange messages had been received already, ignore the message just received.
	if state.EnoughViewChanges() {
		orderer.logger.Log(logging.LevelDebug, "Ignoring ViewChange message, have enough already", "from", from)
		return
	}

	// Update the view change state by the received ViewChange message.
	state.AddSignedViewChange(svc, from)

	orderer.logger.Log(logging.LevelDebug, "Added ViewChange.", "numViewChanges", len(state.signedViewChanges))

	// If enough ViewChange messages have been received
	if state.EnoughViewChanges() {
		orderer.logger.Log(logging.LevelDebug, "Received enough ViewChanges.")

		// Fill in empty Preprepare messages for all sequence numbers where nothing was prepared in the old view.
		emptyPreprepareData := state.SetEmptyPreprepares(vc.View, orderer.segment.Proposals)

		// Request hashing of the new Preprepare messages
		hasherpbdsl.Request(m,
			orderer.moduleConfig.Hasher,
			emptyPreprepareData,
			&vc.View)
	}

	// TODO: Consider checking whether a quorum of valid view change messages has been almost received
	//       and if yes, sending a ViewChange as well if it is the last one missing.

}

func (orderer *Orderer) applyEmptyPreprepareHashResult(
	m dsl.Module,
	digests [][]byte,
	view ot.ViewNr,
) error {

	// Ignore hash result if the view has advanced in the meantime.
	if view < orderer.view {
		orderer.logger.Log(logging.LevelDebug, "Aborting construction of NewView after hashing. View already advanced.",
			"hashView", view,
			"localView", orderer.view,
		)
		return nil
	}

	// Look up the corresponding view change state.
	// No presence check needed, as the entry must exist.
	state := orderer.viewChangeStates[view]

	// Set the digests of empty Preprepares that have just been computed.
	if err := state.SetEmptyPreprepareDigests(digests); err != nil {
		return fmt.Errorf("error setting empty preprepare digests: %w", err)
	}

	// Check if all preprepare messages that need to be re-proposed are locally present.
	state.SetLocalPreprepares(orderer, view)
	if state.HasAllPreprepares() {
		orderer.logger.Log(logging.LevelDebug, "All Preprepares present, sending NewView.")
		// If we have all preprepares, start view change.
		orderer.sendNewView(m, view, state)
		return nil
	}

	// If some Preprepares for re-proposing are still missing, fetch them from other nodes.
	orderer.logger.Log(logging.LevelDebug, "Some Preprepares missing. Asking for retransmission.")
	state.askForMissingPreprepares(m, orderer.moduleConfig)

	return nil
}

func (orderer *Orderer) applyMsgPreprepareRequest(
	m dsl.Module,
	digest []byte,
	sn tt.SeqNr,
	from t.NodeID,
) {
	if preprepare := orderer.lookUpPreprepare(sn, digest); preprepare != nil {

		// If the requested Preprepare message is available, send it to the originator of the request.
		// No need for periodic re-transmission.
		// In the worst case, dropping of these messages may result in another view change,
		// but will not compromise correctness.
		transportpbdsl.SendMessage(
			m,
			orderer.moduleConfig.Net,
			pbftpbmsgs.MissingPreprepare(orderer.moduleConfig.Self, preprepare),
			[]t.NodeID{from})

	}

	// If the requested Preprepare message is not available, ignore the request.
}

func (orderer *Orderer) applyMsgMissingPreprepare(m dsl.Module, preprepare *pbftpbtypes.Preprepare, _ t.NodeID) {

	// Ignore preprepare if received in the meantime or if view has already advanced.
	// This check is technically redundant, as it is (and must be) performed also after the Preprepare is hashed.
	// However, it might prevent some unnecessary hash computation if performed here as well.
	state, view := orderer.latestPendingVCState()
	if pp, ok := state.preprepares[preprepare.Sn]; (ok && pp != nil) || view < orderer.view {
		return
	}

	// Request a hash of the received preprepare message.
	hasherpbdsl.Request(
		m,
		orderer.moduleConfig.Hasher,
		[]*hasherpbtypes.HashData{serializePreprepareForHashing(preprepare)},
		&pbftpbtypes.MissingPreprepare{Preprepare: preprepare},
	)
}

func (orderer *Orderer) applyMissingPreprepareHashResult(
	m dsl.Module,
	digest []byte,
	preprepare *pbftpbtypes.Preprepare,
) {

	// Convenience variable
	sn := preprepare.Sn

	// Look up the latest (with the highest) pending view change state.
	// (Only the view change states that might be waiting for a missing preprepare are considered.)
	state, view := orderer.latestPendingVCState()

	// Ignore preprepare if received in the meantime or if view has already advanced.
	// (Such a situation can occur if missing Preprepares arrive late.)
	if pp, ok := state.preprepares[preprepare.Sn]; (ok && pp != nil) || view < orderer.view {
		return
	}

	// Add the missing preprepare message if its digest matches, updating its view.
	// Note that copying a preprepare with an updated view preserves its hash.
	if bytes.Equal(state.reproposals[sn], digest) && state.preprepares[sn] == nil {
		state.preprepares[sn] = copyPreprepareToNewView(preprepare, view)
	}

	orderer.logger.Log(logging.LevelDebug, "Received missing Preprepare message.", "sn", sn)

	// If this was the last missing preprepare message, proceed to sending a NewView message.
	if state.HasAllPreprepares() {
		orderer.sendNewView(m, view, state)
	}
}

func (orderer *Orderer) sendNewView(m dsl.Module, view ot.ViewNr, vcState *pbftViewChangeState) {

	orderer.logger.Log(logging.LevelDebug, "Sending NewView.")

	// Extract SignedViewChanges and their senders from the view change state.
	viewChangeSenders := make([]t.NodeID, 0, len(vcState.signedViewChanges))
	signedViewChanges := make([]*pbftpbtypes.SignedViewChange, 0, len(vcState.signedViewChanges))
	maputil.IterateSorted(
		vcState.signedViewChanges,
		func(sender t.NodeID, signedViewChange *pbftpbtypes.SignedViewChange) bool {
			viewChangeSenders = append(viewChangeSenders, sender)
			signedViewChanges = append(signedViewChanges, signedViewChange)
			return true
		},
	)

	// Extract re-proposed Preprepares and their corresponding sequence numbers from the view change state.
	preprepareSeqNrs := make([]tt.SeqNr, 0, len(vcState.preprepares))
	preprepares := make([]*pbftpbtypes.Preprepare, 0, len(vcState.preprepares))
	maputil.IterateSorted(vcState.preprepares, func(sn tt.SeqNr, preprepare *pbftpbtypes.Preprepare) bool {
		preprepareSeqNrs = append(preprepareSeqNrs, sn)
		preprepares = append(preprepares, preprepare)
		return true
	})

	// Construct and send the NewView message.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	transportpbdsl.SendMessage(
		m,
		orderer.moduleConfig.Net,
		pbftpbmsgs.NewView(
			orderer.moduleConfig.Self,
			view,
			t.NodeIDSlicePb(viewChangeSenders),
			signedViewChanges,
			preprepareSeqNrs,
			preprepares,
		),
		orderer.segment.NodeIDs(),
	)
}

func (orderer *Orderer) applyMsgNewView(m dsl.Module, newView *pbftpbtypes.NewView, from t.NodeID) {

	// Ignore message if the sender is not the primary of the view.
	if from != orderer.segment.PrimaryNode(newView.View) {
		return
	}

	// Assemble request for checking signatures on the contained ViewChange messages.
	viewChangeData := make([]*cryptopbtypes.SignedData, len(newView.SignedViewChanges))
	signatures := make([][]byte, len(newView.SignedViewChanges))
	for i, signedViewChange := range newView.SignedViewChanges {
		viewChangeData[i] = serializeViewChangeForSigning(signedViewChange.ViewChange)
		signatures[i] = signedViewChange.Signature
	}

	// Request checking of signatures on the contained ViewChange messages
	cryptopbdsl.VerifySigs(
		m,
		orderer.moduleConfig.Crypto,
		viewChangeData,
		signatures,
		t.NodeIDSlice(newView.ViewChangeSenders),
		newView,
	)

}

func (orderer *Orderer) applyVerifiedNewView(m dsl.Module, newView *pbftpbtypes.NewView) {
	// Serialize obtained Preprepare messages for hashing.
	dataToHash := make([]*hasherpbtypes.HashData, len(newView.Preprepares))
	for i, preprepare := range newView.Preprepares { // Preprepares in a NewView message are sorted by sequence number.
		dataToHash[i] = serializePreprepareForHashing(preprepare)
	}

	// Request hashes of the Preprepare messages.
	hasherpbdsl.Request(
		m,
		orderer.moduleConfig.Hasher,
		dataToHash,
		newView,
	)
}

func (orderer *Orderer) applyNewViewHashResult(m dsl.Module, digests [][]byte, newView *pbftpbtypes.NewView) error {

	// Convenience variable
	msgView := newView.View

	// Ignore message if old.
	if msgView < orderer.view {
		return nil
	}

	// Create a temporary view change state object
	// to use for reconstructing the re-proposals from the obtained view change messages.
	vcState := newPbftViewChangeState(orderer.segment.SeqNrs(), orderer.segment.NodeIDs(), orderer.logger)

	// Feed all obtained ViewChange messages to the view change state.
	for i, signedViewChange := range newView.SignedViewChanges {
		vcState.AddSignedViewChange(signedViewChange, t.NodeID(newView.ViewChangeSenders[i]))
	}

	// If the obtained ViewChange messages are not sufficient to infer all re-proposals, ignore NewView message.
	if !vcState.EnoughViewChanges() {
		return nil
	}

	// Verify if the re-proposed hashes match the obtained Preprepares.
	i := 0
	prepreparesMatching := true
	maputil.IterateSorted(vcState.reproposals, func(sn tt.SeqNr, digest []byte) (cont bool) {

		// If the expected digest is empty, it means that the corresponding Preprepare is an "aborted" one.
		// In this case, check the Preprepare directly.
		if len(digest) == 0 {
			prepreparesMatching = validFreshPreprepare(newView.Preprepares[i], msgView, sn)
		} else {
			prepreparesMatching = bytes.Equal(digest, digests[i])
		}

		i++
		return prepreparesMatching
	})

	// If the NewView contains mismatching Preprepares, ignore the message.
	if !prepreparesMatching {
		orderer.logger.Log(logging.LevelWarn, "Hash mismatch in received NewView. Ignoring.", "view", newView.View)
		return nil
	}

	// If all the checks passed, (TODO: make sure all the checks of the NewView message have been performed!)
	// enter the new view.
	err := orderer.initView(m, msgView)
	if err != nil {
		return err
	}

	// Apply all the Preprepares contained in the NewView
	primary := orderer.segment.PrimaryNode(msgView)
	for _, preprepare := range newView.Preprepares {
		orderer.applyMsgPreprepare(m, preprepare, primary)
	}
	return nil
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
func (orderer *Orderer) getViewChangeState(view ot.ViewNr) *pbftViewChangeState {

	if vcs, ok := orderer.viewChangeStates[view]; ok {
		// If a view change state is already present, return it.
		return vcs
	}

	// If no view change state is yet associated with this view, allocate a new one and return it.
	orderer.viewChangeStates[view] = newPbftViewChangeState(orderer.segment.SeqNrs(), orderer.segment.NodeIDs(), orderer.logger)

	return orderer.viewChangeStates[view]
}

// Returns the view change state with the highest view number that received enough view change messages
// (along with the view number itself).
// If there is no view change state with enough ViewChange messages received, returns nil.
func (orderer *Orderer) latestPendingVCState() (*pbftViewChangeState, ot.ViewNr) {

	// View change state with the highest view that received enough ViewChange messages and its view number.
	var state *pbftViewChangeState
	var view ot.ViewNr

	// Find and return the view change state with the highest view number that received enough ViewChange messages.
	for v, s := range orderer.viewChangeStates {
		if s.EnoughViewChanges() && (state == nil || v > view) {
			state, view = s, v
		}
	}
	return state, view
}

// ============================================================
// ViewChange message construction
// ============================================================

// viewChangePSet represents the P set of a PBFT view change message.
// For each sequence number, it holds the digest of the last prepared value,
// along with the view in which it was prepared.
type viewChangePSet map[tt.SeqNr]*pbftpbtypes.PSetEntry

// PbType returns a protobuf type representation (not a raw protobuf, but the generated type) of a viewChangePSet,
// Where all entries are stored in a simple list.
// The list is sorted for repeatability.
func (pSet viewChangePSet) PbType() []*pbftpbtypes.PSetEntry {

	list := make([]*pbftpbtypes.PSetEntry, 0, len(pSet))

	for _, pEntry := range pSet {
		list = append(list, pEntry)
	}

	sort.Slice(list, func(i int, j int) bool {
		if list[i].Sn != list[j].Sn {
			return list[i].Sn < list[j].Sn
		} else if list[i].View != list[j].View {
			return list[i].View < list[j].View
		} else {
			return bytes.Compare(list[i].Digest, list[j].Digest) < 0
		}
	})

	return list
}

func reconstructPSet(entries []*pbftpbtypes.PSetEntry) (viewChangePSet, error) {
	pSet := make(viewChangePSet)
	for _, entry := range entries {

		// There can be at most one entry per sequence number. Otherwise, the set is not valid.
		if _, ok := pSet[entry.Sn]; ok {
			return nil, fmt.Errorf("invalid Pset: conflicting prepare entries")
		}

		pSet[entry.Sn] = entry
	}

	return pSet, nil
}

// The Q set of a PBFT view change message.
// For each sequence number, it holds the digests (encoded as string map keys)
// of all values preprepared for that sequence number,
// along with the latest view in which each of them was preprepared.
type viewChangeQSet map[tt.SeqNr]map[string]ot.ViewNr

// PbType returns a protobuf tye representation (not a raw protobuf, but the generated type) of a viewChangeQSet,
// where all entries, represented as (sn, view, digest) tuples, are stored in a simple list.
// The list is sorted for repeatability.
func (qSet viewChangeQSet) PbType() []*pbftpbtypes.QSetEntry {

	list := make([]*pbftpbtypes.QSetEntry, 0, len(qSet))

	for sn, qEntry := range qSet {
		for digest, view := range qEntry {
			list = append(list, &pbftpbtypes.QSetEntry{
				Sn:     sn,
				View:   view,
				Digest: []byte(digest),
			})
		}
	}

	sort.Slice(list, func(i int, j int) bool {
		if list[i].Sn != list[j].Sn {
			return list[i].Sn < list[j].Sn
		} else if list[i].View != list[j].View {
			return list[i].View < list[j].View
		} else {
			return bytes.Compare(list[i].Digest, list[j].Digest) < 0
		}
	})

	return list
}

func reconstructQSet(entries []*pbftpbtypes.QSetEntry) (viewChangeQSet, error) {
	qSet := make(viewChangeQSet)
	for _, entry := range entries {

		var snEntry map[string]ot.ViewNr
		if sne, ok := qSet[entry.Sn]; ok {
			snEntry = sne
		} else {
			snEntry = make(map[string]ot.ViewNr)
			qSet[entry.Sn] = snEntry
		}

		// There can be at most one entry per digest and sequence number. Otherwise, the set is not valid.
		if _, ok := snEntry[string(entry.Digest)]; ok {
			return nil, fmt.Errorf("invalid Qset: conflicting preprepare entries")
		}

		snEntry[string(entry.Digest)] = entry.View

	}

	return qSet, nil

}

func reconstructPSetQSet(
	signedViewChanges map[t.NodeID]*pbftpbtypes.SignedViewChange,
	logger logging.Logger,
) (map[t.NodeID]viewChangePSet, map[t.NodeID]viewChangeQSet) {
	pSets := make(map[t.NodeID]viewChangePSet)
	qSets := make(map[t.NodeID]viewChangeQSet)

	for nodeID, svc := range signedViewChanges {
		var err error
		var pSet viewChangePSet
		var qSet viewChangeQSet

		pSet, err = reconstructPSet(svc.ViewChange.PSet)
		if err != nil {
			logger.Log(logging.LevelWarn, "could not reconstruct PSet for PBFT view change", "err", err)
			continue
		}

		qSet, err = reconstructQSet(svc.ViewChange.QSet)
		if err != nil {
			logger.Log(logging.LevelWarn, "could not reconstruct QSet for PBFT view change", "err", err)
			continue
		}

		pSets[nodeID] = pSet
		qSets[nodeID] = qSet
	}

	return pSets, qSets
}

// getPSetQSet computes the P set and Q set for the construction of a PBFT view change message.
// Note that this representation of the PSet and QSet is internal to the protocol implementation
// and cannot be directly used in a view change message.
// They must first be transformed to a serializable representation that adheres to the message format.
func (orderer *Orderer) getPSetQSet() (pSet viewChangePSet, qSet viewChangeQSet) {
	// Initialize the PSet.
	pSet = make(map[tt.SeqNr]*pbftpbtypes.PSetEntry)

	// Initialize the QSet.
	qSet = make(map[tt.SeqNr]map[string]ot.ViewNr)

	// For each sequence number, compute the PSet and the QSet.
	for _, sn := range orderer.segment.SeqNrs() {

		// Initialize QSet.
		// (No need to initialize the PSet, as, unlike the PSet,
		// the QSet may hold multiple values for the same sequence number.)
		qSet[sn] = make(map[string]ot.ViewNr)

		// Traverse all previous views.
		// The direction of iteration is important, so the values from newer views can overwrite values from older ones.
		for view := ot.ViewNr(0); view < orderer.view; view++ {
			// Skip views that the node did not even enter
			if slots, ok := orderer.slots[view]; ok {

				// Get the pbftSlot of sn in view (convenience variable)
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

// ============================================================
// NewView message construction
// ============================================================

func reproposal(
	pSets map[t.NodeID]viewChangePSet,
	qSets map[t.NodeID]viewChangeQSet,
	sn tt.SeqNr,
	numNodes int,
) ([]byte, []t.NodeID) {

	if nothingPreparedB(pSets, sn, numNodes) {

		return []byte{}, nil

	}

	for _, pSet := range pSets {
		if entry, ok := pSet[sn]; ok {
			a2, prepreparedIDs := enoughPrepreparesA2(qSets, sn, entry.Digest, entry.View, numNodes)
			if noPrepareConflictsA1(pSets, sn, entry.Digest, entry.View, numNodes) && a2 {

				return entry.Digest, prepreparedIDs

			}
		}
	}

	return nil, nil
}

func noPrepareConflictsA1(
	pSets map[t.NodeID]viewChangePSet,
	sn tt.SeqNr,
	digest []byte,
	view ot.ViewNr,
	numNodes int,
) bool {
	numNonConflicting := 0

	for _, pSet := range pSets {
		if entry, ok := pSet[sn]; !ok {
			numNonConflicting++
		} else {
			if entry.View < view ||
				(entry.View == view && bytes.Equal(entry.Digest, digest)) {
				numNonConflicting++
			}
		}
	}

	return numNonConflicting >= config.StrongQuorum(numNodes)
}

func enoughPrepreparesA2(
	qSets map[t.NodeID]viewChangeQSet,
	sn tt.SeqNr,
	digest []byte,
	view ot.ViewNr,
	numNodes int,
) (bool, []t.NodeID) {

	numPrepares := 0
	nodeIDs := make([]t.NodeID, 0, numNodes)

	for nodeID, qSet := range qSets {
		if snEntry, ok := qSet[sn]; ok {
			if snEntry[string(digest)] >= view {
				numPrepares++
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
	}

	return numPrepares >= config.WeakQuorum(numNodes), nodeIDs
}

func nothingPreparedB(pSets map[t.NodeID]viewChangePSet, sn tt.SeqNr, numNodes int) bool {
	nothingPrepared := 0

	for _, pSet := range pSets {
		if _, ok := pSet[sn]; !ok {
			nothingPrepared++
		}
	}

	return nothingPrepared >= config.StrongQuorum(numNodes)
}
