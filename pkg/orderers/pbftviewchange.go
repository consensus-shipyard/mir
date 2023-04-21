package orderers

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/filecoin-project/mir/pkg/iss/config"
	types2 "github.com/filecoin-project/mir/pkg/orderers/types"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	cryptopbevents "github.com/filecoin-project/mir/pkg/pb/cryptopb/events"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	eventpbevents "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	hasherpbevents "github.com/filecoin-project/mir/pkg/pb/hasherpb/events"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	timertypes "github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// Creates a new Preprepare message identical to the one given as argument,
// except for the view number being set to view.
// The Preprepare message produced by this function has the same digest as the original preprepare,
// since the view number is not used for hash computation.
func copyPreprepareToNewView(preprepare *pbftpbtypes.Preprepare, view types2.ViewNr) *pbftpbtypes.Preprepare {
	return &pbftpbtypes.Preprepare{preprepare.Sn, view, preprepare.Data, preprepare.Aborted} // nolint:govet
}

// ============================================================
// OrdererEvent handling
// ============================================================

// applyViewChangeSNTimeout applies the view change SN timeout event
// triggered some time after a certificate is committed.
// If nothing has been committed since, triggers a view change.
func (orderer *Orderer) applyViewChangeSNTimeout(timeoutEvent *pbftpb.VCSNTimeout) *events.EventList {

	// If the view is still the same as when the timer was set up,
	// if nothing has been committed since then, and if the segment-level checkpoint is not yet stable
	if types2.ViewNr(timeoutEvent.View) == orderer.view &&
		int(timeoutEvent.NumCommitted) == orderer.numCommitted(orderer.view) &&
		!orderer.segmentCheckpoint.Stable(len(orderer.segment.Membership.Nodes)) {

		// Start the view change sub-protocol.
		orderer.logger.Log(logging.LevelWarn, "View change SN timer expired.",
			"view", orderer.view, "numCommitted", timeoutEvent.NumCommitted)
		return orderer.startViewChange()
	}

	// Do nothing otherwise.
	return events.EmptyList()
}

// applyViewChangeSegmentTimeout applies the view change segment timeout event
// triggered some time after a segment is initialized.
// If not all slots have been committed, and the view has not advanced, triggers a view change.
func (orderer *Orderer) applyViewChangeSegmentTimeout(view types2.ViewNr) *events.EventList {

	// TODO: All slots being committed is not sufficient to stop view changes.
	//       An instance-local stable checkpoint must be created as well.

	// If the view is still the same as when the timer was set up and the segment-level checkpoint is not yet stable
	if view == orderer.view && !orderer.segmentCheckpoint.Stable(len(orderer.segment.Membership.Nodes)) {
		// Start the view change sub-protocol.
		orderer.logger.Log(logging.LevelWarn, "View change segment timer expired.", "view", orderer.view)
		return orderer.startViewChange()
	}

	// Do nothing otherwise.
	return events.EmptyList()
}

// startViewChange initiates the view change subprotocol.
// It is triggered on expiry of the SN timeout or the segment timeout.
// It constructs the PBFT view change message and creates an event requesting signing it.
func (orderer *Orderer) startViewChange() *events.EventList {

	eventsOut := events.EmptyList()

	// Enter the view change state and initialize a new view
	orderer.inViewChange = true
	eventsOut.PushBackList(orderer.initView(orderer.view + 1))

	// Compute the P set and Q set to be included in the ViewChange message.
	pSet, qSet := orderer.getPSetQSet()

	// Create a new ViewChange message.
	viewChange := pbftpbtypes.ViewChange{orderer.view, pSet.PbType(), qSet.PbType()} // nolint:govet
	//viewChange := pbftViewChangeMsg(orderer.view, pSet, qSet)

	orderer.logger.Log(logging.LevelWarn, "Starting view change.", "view", orderer.view)

	// Request a signature for the newly created ViewChange message.
	// Operation continues on reception of the SignResult event.
	return eventsOut.PushBack(cryptopbevents.SignRequest(
		orderer.moduleConfig.Crypto,
		serializeViewChangeForSigning(&viewChange),
		SignOrigin(orderer.moduleConfig.Self,
			viewChangeSignOrigin(viewChange.Pb())),
	).Pb())
}

// applyViewChangeSignResult processes a newly generated signature of a ViewChange message.
// It creates a SignedViewChange message and sends it to the new leader (PBFT primary)
func (orderer *Orderer) applyViewChangeSignResult(signature []byte, viewChange *pbftpb.ViewChange) *events.EventList {

	// Convenience variable
	msgView := types2.ViewNr(viewChange.View)

	// Compute the primary of this view (using round-robin on the membership)
	primary := orderer.segment.PrimaryNode(msgView)

	// Repeatedly send the ViewChange message. Repeat until this instance of PBFT is garbage-collected,
	// i.e., from the point of view of the PBFT protocol, effectively forever.
	repeatedSendEvent := eventpbevents.TimerRepeat(
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{transportpbevents.SendMessage(
			orderer.moduleConfig.Net,
			pbftpbmsgs.SignedViewChange(orderer.moduleConfig.Self, pbftpbtypes.ViewChangeFromPb(viewChange), signature),
			[]t.NodeID{primary},
		)},
		timertypes.Duration(orderer.config.ViewChangeResendPeriod),
		tt.RetentionIndex(orderer.config.epochNr),
	)
	return events.ListOf(repeatedSendEvent.Pb())
}

// applyMsgSignedViewChange applies a signed view change message.
// The only thing it does is request verification of the signature.
func (orderer *Orderer) applyMsgSignedViewChange(svc *pbftpbtypes.SignedViewChange, from t.NodeID) *events.EventList {
	viewChange := svc.ViewChange
	return events.ListOf(cryptopbevents.VerifySigs(
		orderer.moduleConfig.Crypto,
		[]*cryptopbtypes.SignedData{serializeViewChangeForSigning(viewChange)},
		[][]byte{svc.Signature},
		SigVerOrigin(orderer.moduleConfig.Self, viewChangeSigVerOrigin(svc.Pb())),
		[]t.NodeID{from},
	).Pb())
}

func (orderer *Orderer) applyVerifiedViewChange(svc *pbftpb.SignedViewChange, from t.NodeID) *events.EventList {
	orderer.logger.Log(logging.LevelDebug, "Received ViewChange.", "sender", from)

	// Convenience variables.
	vc := svc.ViewChange
	vcView := types2.ViewNr(vc.View)

	// Ignore message if it is from an old view.
	if vcView < orderer.view {
		orderer.logger.Log(logging.LevelDebug, "Ignoring ViewChange from old view.",
			"sender", from,
			"vcView", vcView,
			"localView", orderer.view,
		)
		return events.EmptyList()
	}

	// Discard ViewChange message if this node is not the primary for the referenced view
	primary := orderer.segment.PrimaryNode(vcView)
	if orderer.ownID != primary {
		orderer.logger.Log(logging.LevelDebug, "Ignoring ViewChange. Not the primary of view",
			"sender", from,
			"vcView", vcView,
			"primary", primary,
		)
		return events.EmptyList()
	}

	// Look up the state associated with the view change sub-protocol.
	state := orderer.getViewChangeState(vcView)

	// If enough ViewChange messages had been received already, ignore the message just received.
	if state.EnoughViewChanges() {
		orderer.logger.Log(logging.LevelDebug, "Ignoring ViewChange message, have enough already", "from", from)
		return events.EmptyList()
	}

	// Update the view change state by the received ViewChange message.
	state.AddSignedViewChange(pbftpbtypes.SignedViewChangeFromPb(svc), from)

	orderer.logger.Log(logging.LevelDebug, "Added ViewChange.", "numViewChanges", len(state.signedViewChanges))

	// If enough ViewChange messages have been received
	if state.EnoughViewChanges() {
		orderer.logger.Log(logging.LevelDebug, "Received enough ViewChanges.")

		// Fill in empty Preprepare messages for all sequence numbers where nothing was prepared in the old view.
		emptyPreprepareData := state.SetEmptyPreprepares(vcView, orderer.segment.Proposals)

		// Request hashing of the new Preprepare messages
		return events.ListOf(hasherpbevents.Request(
			orderer.moduleConfig.Hasher,
			emptyPreprepareData,
			HashOrigin(orderer.moduleConfig.Self, emptyPreprepareHashOrigin(vcView)),
		).Pb())
	}

	// TODO: Consider checking whether a quorum of valid view change messages has been almost received
	//       and if yes, sending a ViewChange as well if it is the last one missing.

	return events.EmptyList()
}

func (orderer *Orderer) applyEmptyPreprepareHashResult(digests [][]byte, view types2.ViewNr) *events.EventList {

	// Ignore hash result if the view has advanced in the meantime.
	if view < orderer.view {
		orderer.logger.Log(logging.LevelDebug, "Aborting construction of NewView after hashing. View already advanced.",
			"hashView", view,
			"localView", orderer.view,
		)
		return events.EmptyList()
	}

	// Look up the corresponding view change state.
	// No presence check needed, as the entry must exist.
	state := orderer.viewChangeStates[view]

	// Set the digests of empty Preprepares that have just been computed.
	state.SetEmptyPreprepareDigests(digests)

	// Check if all preprepare messages that need to be re-proposed are locally present.
	state.SetLocalPreprepares(orderer, view)
	if state.HasAllPreprepares() {
		orderer.logger.Log(logging.LevelDebug, "All Preprepares present, sending NewView.")
		// If we have all preprepares, start view change.
		return orderer.sendNewView(view, state)
	}

	// If some Preprepares for re-proposing are still missing, fetch them from other nodes.
	orderer.logger.Log(logging.LevelDebug, "Some Preprepares missing. Asking for retransmission.")
	return state.askForMissingPreprepares(orderer.moduleConfig)
}

func (orderer *Orderer) applyMsgPreprepareRequest(
	preprepareRequest *pbftpbtypes.PreprepareRequest,
	from t.NodeID,
) *events.EventList {
	if preprepare := orderer.lookUpPreprepare(preprepareRequest.Sn, preprepareRequest.Digest); preprepare != nil {

		// If the requested Preprepare message is available, send it to the originator of the request.
		// No need for periodic re-transmission.
		// In the worst case, dropping of these messages may result in another view change,
		// but will not compromise correctness.
		return events.ListOf(
			transportpbevents.SendMessage(
				orderer.moduleConfig.Net,
				pbftpbmsgs.MissingPreprepare(orderer.moduleConfig.Self, preprepare),
				[]t.NodeID{from}).Pb(),
		)

	}

	// If the requested Preprepare message is not available, ignore the request.
	return events.EmptyList()
}

func (orderer *Orderer) applyMsgMissingPreprepare(preprepare *pbftpbtypes.Preprepare, _ t.NodeID) *events.EventList {

	// Ignore preprepare if received in the meantime or if view has already advanced.
	// This check is technically redundant, as it is (and must be) performed also after the Preprepare is hashed.
	// However, it might prevent some unnecessary hash computation if performed here as well.
	state, view := orderer.latestPendingVCState()
	if pp, ok := state.preprepares[preprepare.Sn]; (ok && pp != nil) || view < orderer.view {
		return events.EmptyList()
	}

	// Request a hash of the received preprepare message.
	return events.ListOf(hasherpbevents.Request(
		orderer.moduleConfig.Hasher,
		[]*commonpbtypes.HashData{serializePreprepareForHashing(preprepare)},
		HashOrigin(orderer.moduleConfig.Self, missingPreprepareHashOrigin(preprepare.Pb())),
	).Pb())
}

func (orderer *Orderer) applyMissingPreprepareHashResult(
	digest []byte,
	preprepare *pbftpbtypes.Preprepare,
) *events.EventList {

	// Convenience variable
	sn := preprepare.Sn

	// Look up the latest (with the highest) pending view change state.
	// (Only the view change states that might be waiting for a missing preprepare are considered.)
	state, view := orderer.latestPendingVCState()

	// Ignore preprepare if received in the meantime or if view has already advanced.
	// (Such a situation can occur if missing Preprepares arrive late.)
	if pp, ok := state.preprepares[preprepare.Sn]; (ok && pp != nil) || view < orderer.view {
		return events.EmptyList()
	}

	// Add the missing preprepare message if its digest matches, updating its view.
	// Note that copying a preprepare with an updated view preserves its hash.
	if bytes.Equal(state.reproposals[sn], digest) && state.preprepares[sn] == nil {
		state.preprepares[sn] = copyPreprepareToNewView(preprepare, view)
	}

	orderer.logger.Log(logging.LevelDebug, "Received missing Preprepare message.", "sn", sn)

	// If this was the last missing preprepare message, proceed to sending a NewView message.
	if state.HasAllPreprepares() {
		return orderer.sendNewView(view, state)
	}

	return events.EmptyList()
}

func (orderer *Orderer) sendNewView(view types2.ViewNr, vcState *pbftViewChangeState) *events.EventList {

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
	return events.ListOf(transportpbevents.SendMessage(
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
	).Pb())
}

func (orderer *Orderer) applyMsgNewView(newView *pbftpbtypes.NewView, from t.NodeID) *events.EventList {

	// Ignore message if the sender is not the primary of the view.
	if from != orderer.segment.PrimaryNode(newView.View) {
		return events.EmptyList()
	}

	// Assemble request for checking signatures on the contained ViewChange messages.
	viewChangeData := make([]*cryptopbtypes.SignedData, len(newView.SignedViewChanges))
	signatures := make([][]byte, len(newView.SignedViewChanges))
	for i, signedViewChange := range newView.SignedViewChanges {
		viewChangeData[i] = serializeViewChangeForSigning(signedViewChange.ViewChange)
		signatures[i] = signedViewChange.Signature
	}

	// Request checking of signatures on the contained ViewChange messages
	return events.ListOf(cryptopbevents.VerifySigs(
		orderer.moduleConfig.Crypto,
		viewChangeData,
		signatures,
		SigVerOrigin(
			orderer.moduleConfig.Self,
			newViewSigVerOrigin(newView.Pb())),
		t.NodeIDSlice(newView.ViewChangeSenders),
	).Pb())
}

func (orderer *Orderer) applyVerifiedNewView(newView *pbftpb.NewView) *events.EventList {
	// Serialize obtained Preprepare messages for hashing.
	dataToHash := make([]*commonpbtypes.HashData, len(newView.Preprepares))
	for i, preprepare := range newView.Preprepares { // Preprepares in a NewView message are sorted by sequence number.
		dataToHash[i] = serializePreprepareForHashing(pbftpbtypes.PreprepareFromPb(preprepare))
	}

	// Request hashes of the Preprepare messages.
	return events.ListOf(hasherpbevents.Request(
		orderer.moduleConfig.Hasher,
		dataToHash,
		HashOrigin(orderer.moduleConfig.Self, newViewHashOrigin(newView)),
	).Pb())
}

func (orderer *Orderer) applyNewViewHashResult(digests [][]byte, newView *pbftpbtypes.NewView) *events.EventList {

	// Convenience variable
	msgView := newView.View

	// Ignore message if old.
	if msgView < orderer.view {
		return events.EmptyList()
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
		return events.EmptyList()
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
		return events.EmptyList()
	}

	// If all the checks passed, (TODO: make sure all the checks of the NewView message have been performed!)
	// enter the new view.
	eventsOut := orderer.initView(msgView)

	// Apply all the Preprepares contained in the NewView
	primary := orderer.segment.PrimaryNode(msgView)
	for _, preprepare := range newView.Preprepares {
		eventsOut.PushBackList(orderer.applyMsgPreprepare(preprepare, primary))
	}
	return eventsOut
}

// ============================================================
// Auxiliary functions
// ============================================================

func validFreshPreprepare(preprepare *pbftpbtypes.Preprepare, view types2.ViewNr, sn tt.SeqNr) bool {
	return preprepare.Aborted &&
		preprepare.Sn == sn &&
		preprepare.View == view
}

// viewChangeState returns the state of the view change sub-protocol associated with the given view,
// allocating the associated data structures as needed.
func (orderer *Orderer) getViewChangeState(view types2.ViewNr) *pbftViewChangeState {

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
func (orderer *Orderer) latestPendingVCState() (*pbftViewChangeState, types2.ViewNr) {

	// View change state with the highest view that received enough ViewChange messages and its view number.
	var state *pbftViewChangeState
	var view types2.ViewNr

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
// For each sequence number, it holds the digest of the last prepared certificate,
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
// of all certificates preprepared for that sequence number,
// along with the latest view in which each of them was preprepared.
type viewChangeQSet map[tt.SeqNr]map[string]types2.ViewNr

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

		var snEntry map[string]types2.ViewNr
		if sne, ok := qSet[entry.Sn]; ok {
			snEntry = sne
		} else {
			snEntry = make(map[string]types2.ViewNr)
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
	qSet = make(map[tt.SeqNr]map[string]types2.ViewNr)

	// For each sequence number, compute the PSet and the QSet.
	for _, sn := range orderer.segment.SeqNrs() {

		// Initialize QSet.
		// (No need to initialize the PSet, as, unlike the PSet,
		// the QSet may hold multiple values for the same sequence number.)
		qSet[sn] = make(map[string]types2.ViewNr)

		// Traverse all previous views.
		// The direction of iteration is important, so the values from newer views can overwrite values from older ones.
		for view := types2.ViewNr(0); view < orderer.view; view++ {
			// Skip views that the node did not even enter
			if slots, ok := orderer.slots[view]; ok {

				// Get the pbftSlot of sn in view (convenience variable)
				slot := slots[sn]

				// If a certificate was prepared for sn in view, add the corresponding entry to the PSet.
				// If there was an entry corresponding to an older view, it will be overwritten.
				if slot.Prepared {
					pSet[sn] = &pbftpbtypes.PSetEntry{
						Sn:     sn,
						View:   view,
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
	view types2.ViewNr,
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
	view types2.ViewNr,
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
