package iss

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/isspbftpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// Creates a new Preprepare message identical to the one given as argument,
// except for the view number being set to view.
// The Preprepare message produced by this function has the same digest as the original preprepare,
// since the view number is not used for hash computation.
func copyPreprepareToNewView(preprepare *isspbftpb.Preprepare, view t.PBFTViewNr) *isspbftpb.Preprepare {
	return pbftPreprepareMsg(t.SeqNr(preprepare.Sn), view, preprepare.CertData, preprepare.Aborted)
}

// ============================================================
// Event handling
// ============================================================

// applyViewChangeSNTimeout applies the view change SN timeout event
// triggered some time after a certificate is committed.
// If nothing has been committed since, triggers a view change.
func (pbft *pbftInstance) applyViewChangeSNTimeout(timeoutEvent *isspbftpb.VCSNTimeout) *events.EventList {

	// If the view is still the same as when the timer was set up,
	// if nothing has been committed since then, and if the segment-level checkpoint is not yet stable
	if t.PBFTViewNr(timeoutEvent.View) == pbft.view &&
		int(timeoutEvent.NumCommitted) == pbft.numCommitted(pbft.view) &&
		!pbft.segmentCheckpoint.Stable(len(pbft.segment.Membership)) {

		// Start the view change sub-protocol.
		pbft.logger.Log(logging.LevelWarn, "View change SN timer expired.",
			"view", pbft.view, "numCommitted", timeoutEvent.NumCommitted)
		return pbft.startViewChange()
	}

	// Do nothing otherwise.
	return events.EmptyList()
}

// applyViewChangeSegmentTimeout applies the view change segment timeout event
// triggered some time after a segment is initialized.
// If not all slots have been committed, and the view has not advanced, triggers a view change.
func (pbft *pbftInstance) applyViewChangeSegmentTimeout(view t.PBFTViewNr) *events.EventList {

	// TODO: All slots being committed is not sufficient to stop view changes.
	//       An instance-local stable checkpoint must be created as well.

	// If the view is still the same as when the timer was set up and the segment-level checkpoint is not yet stable
	if view == pbft.view && !pbft.segmentCheckpoint.Stable(len(pbft.segment.Membership)) {
		// Start the view change sub-protocol.
		pbft.logger.Log(logging.LevelWarn, "View change segment timer expired.", "view", pbft.view)
		return pbft.startViewChange()
	}

	// Do nothing otherwise.
	return events.EmptyList()
}

// startViewChange initiates the view change subprotocol.
// It is triggered on expiry of the SN timeout or the segment timeout.
// It constructs the PBFT view change message and creates an event requesting signing it.
func (pbft *pbftInstance) startViewChange() *events.EventList {

	eventsOut := events.EmptyList()

	// Enter the view change state and initialize a new view
	pbft.inViewChange = true
	eventsOut.PushBackList(pbft.initView(pbft.view + 1))

	// Compute the P set and Q set to be included in the ViewChange message.
	pSet, qSet := pbft.getPSetQSet()

	// Create a new ViewChange message.
	viewChange := pbftViewChangeMsg(pbft.view, pSet, qSet)

	pbft.logger.Log(logging.LevelWarn, "Starting view change.", "view", pbft.view)

	// Request a signature for the newly created ViewChange message.
	// Operation continues on reception of the SignResult event.
	return eventsOut.PushBack(pbft.eventService.SignRequest(
		serializeViewChangeForSigning(viewChange),
		viewChangeSignOrigin(viewChange),
	))
}

// applyViewChangeSignResult processes a newly generated signature of a ViewChange message.
// It creates a SignedViewChange message and sends it to the new leader (PBFT primary)
func (pbft *pbftInstance) applyViewChangeSignResult(signature []byte, viewChange *isspbftpb.ViewChange) *events.EventList {

	// Convenience variable
	msgView := t.PBFTViewNr(viewChange.View)

	// Compute the primary of this view (using round-robin on the membership)
	primary := primaryNode(pbft.segment, msgView)

	// Assemble signature and viewChange to a SignedViewChange message.
	signedViewChange := pbftSignedViewChangeMsg(viewChange, signature)

	// Repeatedly send the ViewChange message. Repeat until this instance of PBFT is garbage-collected,
	// i.e., from the point of view of the PBFT protocol, effectively forever.
	repeatedSendEvent := pbft.eventService.TimerRepeat(
		t.TimeDuration(pbft.config.ViewChangeResendPeriod),
		pbft.eventService.SendMessage(
			PbftSignedViewChangeSBMessage(signedViewChange),
			[]t.NodeID{primary},
		),
	)
	persistEvent := pbft.eventService.WALAppend(PbftPersistSignedViewChange(signedViewChange))
	persistEvent.FollowUp(repeatedSendEvent)
	return events.ListOf(persistEvent)
}

// applyMsgSignedViewChange applies a signed view change message.
// The only thing it does is request verification of the signature.
func (pbft *pbftInstance) applyMsgSignedViewChange(svc *isspbftpb.SignedViewChange, from t.NodeID) *events.EventList {
	viewChange := svc.ViewChange
	return events.ListOf(pbft.eventService.VerifyNodeSigs(
		[][][]byte{serializeViewChangeForSigning(viewChange)},
		[][]byte{svc.Signature},
		[]t.NodeID{from},
		viewChangeSigVerOrigin(svc),
	))
}

func (pbft *pbftInstance) applyVerifiedViewChange(svc *isspbftpb.SignedViewChange, from t.NodeID) *events.EventList {
	pbft.logger.Log(logging.LevelDebug, "Received ViewChange.", "sender", from)

	// Convenience variables.
	vc := svc.ViewChange
	vcView := t.PBFTViewNr(vc.View)

	// Ignore message if it is from an old view.
	if vcView < pbft.view {
		pbft.logger.Log(logging.LevelDebug, "Ignoring ViewChange from old view.",
			"sender", from,
			"vcView", vcView,
			"localView", pbft.view,
		)
		return events.EmptyList()
	}

	// Discard ViewChange message if this node is not the primary for the referenced view
	primary := primaryNode(pbft.segment, vcView)
	if pbft.ownID != primary {
		pbft.logger.Log(logging.LevelDebug, "Ignoring ViewChange. Not the primary of view",
			"sender", from,
			"vcView", vcView,
			"primary", primary,
		)
		return events.EmptyList()
	}

	// Look up the state associated with the view change sub-protocol.
	state := pbft.getViewChangeState(vcView)

	// If enough ViewChange messages had been received already, ignore the message just received.
	if state.EnoughViewChanges() {
		pbft.logger.Log(logging.LevelDebug, "Ignoring ViewChange message, have enough already", "from", from)
		return events.EmptyList()
	}

	// Update the view change state by the received ViewChange message.
	state.AddSignedViewChange(svc, from)

	pbft.logger.Log(logging.LevelDebug, "Added ViewChange.", "numViewChanges", len(state.signedViewChanges))

	// If enough ViewChange messages have been received
	if state.EnoughViewChanges() {
		pbft.logger.Log(logging.LevelDebug, "Received enough ViewChanges.")

		// Fill in empty Preprepare messages for all sequence numbers where nothing was prepared in the old view.
		emptyPreprepareData := state.SetEmptyPreprepares(vcView)

		// Request hashing of the new Preprepare messages
		return events.ListOf(
			pbft.eventService.HashRequest(emptyPreprepareData, emptyPreprepareHashOrigin(vcView)),
		)
	}

	// TODO: Consider checking whether a quorum of valid view change messages has been almost received
	//       and if yes, sending a ViewChange as well if it is the last one missing.

	return events.EmptyList()
}

func (pbft *pbftInstance) applyEmptyPreprepareHashResult(digests [][]byte, view t.PBFTViewNr) *events.EventList {

	// Ignore hash result if the view has advanced in the meantime.
	if view < pbft.view {
		pbft.logger.Log(logging.LevelDebug, "Aborting construction of NewView after hashing. View already advanced.",
			"hashView", view,
			"localView", pbft.view,
		)
		return events.EmptyList()
	}

	// Look up the corresponding view change state.
	// No presence check needed, as the entry must exist.
	state := pbft.viewChangeStates[view]

	// Set the digests of empty Preprepares that have just been computed.
	state.SetEmptyPreprepareDigests(digests)

	// Check if all preprepare messages that need to be re-proposed are locally present.
	state.SetLocalPreprepares(pbft, view)
	if state.HasAllPreprepares() {
		pbft.logger.Log(logging.LevelDebug, "All Preprepares present, sending NewView.")
		// If we have all preprepares, start view change.
		return pbft.sendNewView(view, state)
	}

	// If some Preprepares for re-proposing are still missing, fetch them from other nodes.
	pbft.logger.Log(logging.LevelDebug, "Some Preprepares missing. Asking for retransmission.")
	return state.askForMissingPreprepares(pbft.eventService)
}

func (pbft *pbftInstance) applyMsgPreprepareRequest(
	preprepareRequest *isspbftpb.PreprepareRequest,
	from t.NodeID,
) *events.EventList {
	if preprepare := pbft.lookUpPreprepare(t.SeqNr(preprepareRequest.Sn), preprepareRequest.Digest); preprepare != nil {

		// If the requested Preprepare message is available, send it to the originator of the request.
		// No need for periodic re-transmission.
		// In the worst case, dropping of these messages may result in another view change,
		// but will not compromise correctness.
		return events.ListOf(
			pbft.eventService.SendMessage(PbftMissingPreprepareSBMessage(preprepare), []t.NodeID{from}),
		)

	}

	// If the requested Preprepare message is not available, ignore the request.
	return events.EmptyList()
}

func (pbft *pbftInstance) applyMsgMissingPreprepare(preprepare *isspbftpb.Preprepare, _ t.NodeID) *events.EventList {

	// Ignore preprepare if received in the meantime or if view has already advanced.
	// This check is technically redundant, as it is (and must be) performed also after the Preprepare is hashed.
	// However, it might prevent some unnecessary hash computation if performed here as well.
	state, view := pbft.latestPendingVCState()
	if pp, ok := state.preprepares[t.SeqNr(preprepare.Sn)]; (ok && pp != nil) || view < pbft.view {
		return events.EmptyList()
	}

	// Request a hash of the received preprepare message.
	hashRequest := pbft.eventService.HashRequest(
		[][][]byte{serializePreprepareForHashing(preprepare)},
		missingPreprepareHashOrigin(preprepare),
	)
	return events.ListOf(hashRequest)
}

func (pbft *pbftInstance) applyMissingPreprepareHashResult(
	digest []byte,
	preprepare *isspbftpb.Preprepare,
) *events.EventList {

	// Convenience variable
	sn := t.SeqNr(preprepare.Sn)

	// Look up the latest (with the highest) pending view change state.
	// (Only the view change states that might be waiting for a missing preprepare are considered.)
	state, view := pbft.latestPendingVCState()

	// Ignore preprepare if received in the meantime or if view has already advanced.
	// (Such a situation can occur if missing Preprepares arrive late.)
	if pp, ok := state.preprepares[t.SeqNr(preprepare.Sn)]; (ok && pp != nil) || view < pbft.view {
		return events.EmptyList()
	}

	// Add the missing preprepare message if its digest matches, updating its view.
	// Note that copying a preprepare with an updated view preserves its hash.
	if bytes.Equal(state.reproposals[sn], digest) && state.preprepares[sn] == nil {
		state.preprepares[sn] = copyPreprepareToNewView(preprepare, view)
	}

	pbft.logger.Log(logging.LevelDebug, "Received missing Preprepare message.", "sn", sn)

	// If this was the last missing preprepare message, proceed to sending a NewView message.
	if state.HasAllPreprepares() {
		return pbft.sendNewView(view, state)
	}

	return events.EmptyList()
}

func (pbft *pbftInstance) sendNewView(view t.PBFTViewNr, vcState *pbftViewChangeState) *events.EventList {

	pbft.logger.Log(logging.LevelDebug, "Sending NewView.")

	// Extract SignedViewChanges and their senders from the view change state.
	viewChangeSenders := make([]t.NodeID, 0, len(vcState.signedViewChanges))
	signedViewChanges := make([]*isspbftpb.SignedViewChange, 0, len(vcState.signedViewChanges))
	maputil.IterateSorted(
		vcState.signedViewChanges,
		func(sender t.NodeID, signedViewChange *isspbftpb.SignedViewChange) bool {
			viewChangeSenders = append(viewChangeSenders, sender)
			signedViewChanges = append(signedViewChanges, signedViewChange)
			return true
		},
	)

	// Extract re-proposed Preprepares and their corresponding sequence numbers from the view change state.
	preprepareSeqNrs := make([]t.SeqNr, 0, len(vcState.preprepares))
	preprepares := make([]*isspbftpb.Preprepare, 0, len(vcState.preprepares))
	maputil.IterateSorted(vcState.preprepares, func(sn t.SeqNr, preprepare *isspbftpb.Preprepare) bool {
		preprepareSeqNrs = append(preprepareSeqNrs, sn)
		preprepares = append(preprepares, preprepare)
		return true
	})

	// Construct, persist and send the NewView message.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	newView := pbftNewViewMsg(view, viewChangeSenders, signedViewChanges, preprepareSeqNrs, preprepares)
	persistEvent := pbft.eventService.WALAppend(PbftPersistNewView(newView))
	persistEvent.FollowUp(pbft.eventService.SendMessage(PbftNewViewSBMessage(newView), pbft.segment.Membership))
	return events.ListOf(persistEvent)
}

func (pbft *pbftInstance) applyMsgNewView(newView *isspbftpb.NewView, from t.NodeID) *events.EventList {

	// Ignore message if the sender is not the primary of the view.
	if from != primaryNode(pbft.segment, t.PBFTViewNr(newView.View)) {
		return events.EmptyList()
	}

	// Assemble request for checking signatures on the contained ViewChange messages.
	viewChangeData := make([][][]byte, len(newView.SignedViewChanges))
	signatures := make([][]byte, len(newView.SignedViewChanges))
	for i, signedViewChange := range newView.SignedViewChanges {
		viewChangeData[i] = serializeViewChangeForSigning(signedViewChange.ViewChange)
		signatures[i] = signedViewChange.Signature
	}

	// Request checking of signatures on the contained ViewChange messages
	return events.ListOf(pbft.eventService.VerifyNodeSigs(
		viewChangeData,
		signatures,
		t.NodeIDSlice(newView.ViewChangeSenders),
		newViewSigVerOrigin(newView),
	))
}

func (pbft *pbftInstance) applyVerifiedNewView(newView *isspbftpb.NewView) *events.EventList {
	// Serialize obtained Preprepare messages for hashing.
	dataToHash := make([][][]byte, len(newView.Preprepares))
	for i, preprepare := range newView.Preprepares { // Preprepares in a NewView message are sorted by sequence number.
		dataToHash[i] = serializePreprepareForHashing(preprepare)
	}

	// Request hashes of the Preprepare messages.
	return events.ListOf(pbft.eventService.HashRequest(dataToHash, newViewHashOrigin(newView)))
}

func (pbft *pbftInstance) applyNewViewHashResult(digests [][]byte, newView *isspbftpb.NewView) *events.EventList {

	// Convenience variable
	msgView := t.PBFTViewNr(newView.View)

	// Ignore message if old.
	if msgView < pbft.view {
		return events.EmptyList()
	}

	// Create a temporary view change state object
	// to use for reconstructing the re-proposals from the obtained view change messages.
	vcState := newPbftViewChangeState(pbft.segment.SeqNrs, pbft.segment.Membership, pbft.logger)

	// Feed all obtained ViewChange messages to the view chnage state.
	for i, signedViewChange := range newView.SignedViewChanges {
		vcState.AddSignedViewChange(signedViewChange, t.NodeID(newView.ViewChangeSenders[i]))
	}

	// If the obtained ViewChange messages are not sufficient to infer all re-proposals, ignore NewView message.
	if !vcState.EnoughViewChanges() {
		return events.EmptyList()
	}

	// Verify if the re-proposed hashes match the obtained Preprepares.
	i := 0
	prepreparesMatching := true
	maputil.IterateSorted(vcState.reproposals, func(sn t.SeqNr, digest []byte) (cont bool) {

		// If the expected digest is empty, it means that the corresponding Preprepare is an "aborted" one.
		// In this case, check the Preprepare directly.
		if len(digest) == 0 {
			prepreparesMatching = validEmptyPreprepare(newView.Preprepares[i], msgView, sn)
		} else {
			prepreparesMatching = bytes.Equal(digest, digests[i])
		}

		i++
		return prepreparesMatching
	})

	// If the NewVeiw contains mismatching Preprepares, ignore the message.
	if !prepreparesMatching {
		pbft.logger.Log(logging.LevelWarn, "Hash mismatch in received NewView. Ignoring.", "view", newView.View)
		return events.EmptyList()
	}

	// If all the checks passed, (TODO: make sure all the checks of the NewView message have been performed!)
	// enter the new view.
	eventsOut := pbft.initView(msgView)

	// Apply all the Preprepares contained in the NewView
	primary := primaryNode(pbft.segment, msgView)
	for _, preprepare := range newView.Preprepares {
		eventsOut.PushBackList(pbft.applyMsgPreprepare(preprepare, primary))
	}
	return eventsOut
}

// ============================================================
// Auxiliary functions
// ============================================================

func validEmptyPreprepare(preprepare *isspbftpb.Preprepare, view t.PBFTViewNr, sn t.SeqNr) bool {
	return preprepare.Aborted &&
		t.SeqNr(preprepare.Sn) == sn &&
		t.PBFTViewNr(preprepare.View) == view &&
		len(preprepare.CertData) == 0
}

// viewChangeState returns the state of the view change sub-protocol associated with the given view,
// allocating the associated data structures as needed.
func (pbft *pbftInstance) getViewChangeState(view t.PBFTViewNr) *pbftViewChangeState {

	if vcs, ok := pbft.viewChangeStates[view]; ok {
		// If a view change state is already present, return it.
		return vcs
	}

	// If no view change state is yet associated with this view, allocate a new one and return it.
	pbft.viewChangeStates[view] = newPbftViewChangeState(pbft.segment.SeqNrs, pbft.segment.Membership, pbft.logger)

	return pbft.viewChangeStates[view]
}

// Returns the view change state with the highest view number that received enough view change messages
// (along with the view number itself).
// If there is no view change state with enough ViewChange messages received, returns nil.
func (pbft *pbftInstance) latestPendingVCState() (*pbftViewChangeState, t.PBFTViewNr) {

	// View change state with the highest view that received enough ViewChange messages and its view number.
	var state *pbftViewChangeState
	var view t.PBFTViewNr

	// Find and return the view change state with the highest view number that received enough ViewChange messages.
	for v, s := range pbft.viewChangeStates {
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
// For each sequence number, it holds the digest of the last prepared certificate,
// along with the view in which it was prepared.
type viewChangePSet map[t.SeqNr]*isspbftpb.PSetEntry

// Pb returns a protobuf representation of a viewChangePSet,
// Where all entries are stored in a simple list.
// The list is sorted for repeatability.
func (pSet viewChangePSet) Pb() []*isspbftpb.PSetEntry {

	list := make([]*isspbftpb.PSetEntry, 0, len(pSet))

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

func reconstructPSet(entries []*isspbftpb.PSetEntry) (viewChangePSet, error) {
	pSet := make(viewChangePSet)
	for _, entry := range entries {

		// There can be at most one entry per sequence number. Otherwise, the set is not valid.
		if _, ok := pSet[t.SeqNr(entry.Sn)]; ok {
			return nil, fmt.Errorf("invalid Pset: conflicting prepare entries")
		}

		pSet[t.SeqNr(entry.Sn)] = entry
	}

	return pSet, nil
}

// The Q set of a PBFT view change message.
// For each sequence number, it holds the digests (encoded as string map keys)
// of all certificates preprepared for that sequence number,
// along with the latest view in which each of them was preprepared.
type viewChangeQSet map[t.SeqNr]map[string]t.PBFTViewNr

// Pb returns a protobuf representation of a viewChangeQSet,
// where all entries, represented as (sn, view, digest) tuples, are stored in a simple list.
// The list is sorted for repeatability.
func (qSet viewChangeQSet) Pb() []*isspbftpb.QSetEntry {

	list := make([]*isspbftpb.QSetEntry, 0, len(qSet))

	for sn, qEntry := range qSet {
		for digest, view := range qEntry {
			list = append(list, &isspbftpb.QSetEntry{
				Sn:     sn.Pb(),
				View:   view.Pb(),
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

func reconstructQSet(entries []*isspbftpb.QSetEntry) (viewChangeQSet, error) {
	qSet := make(viewChangeQSet)
	for _, entry := range entries {

		var snEntry map[string]t.PBFTViewNr
		if sne, ok := qSet[t.SeqNr(entry.Sn)]; ok {
			snEntry = sne
		} else {
			snEntry = make(map[string]t.PBFTViewNr)
			qSet[t.SeqNr(entry.Sn)] = snEntry
		}

		// There can be at most one entry per digest and sequence number. Otherwise, the set is not valid.
		if _, ok := snEntry[string(entry.Digest)]; ok {
			return nil, fmt.Errorf("invalid Qset: conflicting preprepare entries")
		}

		snEntry[string(entry.Digest)] = t.PBFTViewNr(entry.View)

	}

	return qSet, nil

}

func reconstructPSetQSet(
	signedViewChanges map[t.NodeID]*isspbftpb.SignedViewChange,
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
func (pbft *pbftInstance) getPSetQSet() (pSet viewChangePSet, qSet viewChangeQSet) {
	// Initialize the PSet.
	pSet = make(map[t.SeqNr]*isspbftpb.PSetEntry)

	// Initialize the QSet.
	qSet = make(map[t.SeqNr]map[string]t.PBFTViewNr)

	// For each sequence number, compute the PSet and the QSet.
	for _, sn := range pbft.segment.SeqNrs {

		// Initialize QSet.
		// (No need to initialize the PSet, as, unlike the PSet,
		// the QSet may hold multiple values for the same sequence number.)
		qSet[sn] = make(map[string]t.PBFTViewNr)

		// Traverse all previous views.
		// The direction of iteration is important, so the values from newer views can overwrite values from older ones.
		for view := t.PBFTViewNr(0); view < pbft.view; view++ {
			// Skip views that the node did not even enter
			if slots, ok := pbft.slots[view]; ok {

				// Get the pbftSlot of sn in view (convenience variable)
				slot := slots[sn]

				// If a certificate was prepared for sn in view, add the corresponding entry to the PSet.
				// If there was an entry corresponding to an older view, it will be overwritten.
				if slot.Prepared {
					pSet[sn] = &isspbftpb.PSetEntry{
						Sn:     sn.Pb(),
						View:   view.Pb(),
						Digest: slot.Digest,
					}
				}

				// If a certificate was preprepared for sn in view, add the corresponding entry to the QSet.
				// If the same certificate has been preprepared in an older view, its entry will be overwritten.
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
	sn t.SeqNr,
	numNodes int,
) ([]byte, []t.NodeID) {

	if nothingPreparedB(pSets, sn, numNodes) {

		return []byte{}, nil

	}

	for _, pSet := range pSets {
		if entry, ok := pSet[sn]; ok {
			a2, prepreparedIDs := enoughPrepreparesA2(qSets, sn, entry.Digest, t.PBFTViewNr(entry.View), numNodes)
			if noPrepareConflictsA1(pSets, sn, entry.Digest, t.PBFTViewNr(entry.View), numNodes) && a2 {

				return entry.Digest, prepreparedIDs

			}
		}
	}

	return nil, nil
}

func noPrepareConflictsA1(
	pSets map[t.NodeID]viewChangePSet,
	sn t.SeqNr,
	digest []byte,
	view t.PBFTViewNr,
	numNodes int,
) bool {
	numNonConflicting := 0

	for _, pSet := range pSets {
		if entry, ok := pSet[sn]; !ok {
			numNonConflicting++
		} else {
			if t.PBFTViewNr(entry.View) < view ||
				(t.PBFTViewNr(entry.View) == view && bytes.Equal(entry.Digest, digest)) {
				numNonConflicting++
			}
		}
	}

	return numNonConflicting >= strongQuorum(numNodes)
}

func enoughPrepreparesA2(
	qSets map[t.NodeID]viewChangeQSet,
	sn t.SeqNr,
	digest []byte,
	view t.PBFTViewNr,
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

	return numPrepares >= weakQuorum(numNodes), nodeIDs
}

func nothingPreparedB(pSets map[t.NodeID]viewChangePSet, sn t.SeqNr, numNodes int) bool {
	nothingPrepared := 0

	for _, pSet := range pSets {
		if _, ok := pSet[sn]; !ok {
			nothingPrepared++
		}
	}

	return nothingPrepared >= strongQuorum(numNodes)
}
