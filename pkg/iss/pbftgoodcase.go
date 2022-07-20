package iss

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/isspbftpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// Data structures and functions
// ============================================================

// pbftProposalState tracks the state of the pbftInstance related to proposing batches.
// The proposal state is only used if this node is the leader of this instance of PBFT.
type pbftProposalState struct {

	// Tracks the number of proposals already made.
	// Used to calculate the sequence number of the next proposal (from the associated segment's sequence numbers)
	// and to stop proposing when a proposal has been made for all sequence numbers.
	proposalsMade int

	// Saves the number of pending requests that are available to be included in the next proposal,
	// as reported by ISS through the PendingRequests event.
	// When this value exceeds config.MaxBatchSize, a proposal might be made before the config.MaxProposeDelay.
	numPendingRequests t.NumRequests

	// Flag indicating whether a new batch has been requested from ISS.
	// This effectively means that a proposal is in progress and no new proposals should be started.
	// This is required, since the PBFT implementation does not assemble request batches itself,
	// as it is not managing the request buckets
	// (which, in general, do not exclusively concern the orderer and are thus managed by ISS directly).
	// Instead, when the PBFT instance is ready to propose,
	// it requests a new batch from the enclosing ISS implementation to provide the batch.
	batchRequested bool

	// If batchRequested is true, batchRequestedView indicates the PBFT view
	// in which the PBFT protocol was when the batch was requested.
	// When the batch is ready, the protocol needs to check whether it is still in the same view
	// as when the batch was requested.
	// If the view advanced in the meantime, the proposal must be aborted and the contained requests resurrected
	// (i.e., put back in their respective ISS bucket queues).
	// If batchRequested is false, batchRequestedView must not be used.
	batchRequestedView t.PBFTViewNr

	// the number of proposals for which the timeout has passed.
	// If this number is higher than proposalsMade, a new batch can be proposed.
	proposalTimeout int
}

// ============================================================
// Event handling
// ============================================================

// applyProposeTimeout applies the event of the proposal timeout firing.
// It updates the proposal state accordingly and triggers a new proposal if possible.
func (pbft *pbftInstance) applyProposeTimeout(numProposals int) *events.EventList {

	// If we are still waiting for this timeout
	if numProposals > pbft.proposal.proposalTimeout {

		// Save the number of proposals for which the timeout fired.
		pbft.proposal.proposalTimeout = numProposals

		// If this was the last bit missing, start a new proposal.
		if pbft.canPropose() {
			return pbft.requestNewBatch()
		}
	}

	return events.EmptyList()
}

// requestNewBatch asks (by means of a CutBatch event) ISS to assemble a new request batch.
// When the batch is ready, it must be passed to the orderer using the BatchReady event.
func (pbft *pbftInstance) requestNewBatch() *events.EventList {

	// Set a flag indicating that a batch has been requested,
	// so that no new batches will be requested before the reception of this one.
	// It will be cleared when BatchReady is received.
	pbft.proposal.batchRequested = true

	// Remember the view in which the batch has been requested
	// to make sure that we are still in the same view when the batch becomes ready.
	pbft.proposal.batchRequestedView = pbft.view

	// Emit the CutBatch event.
	// Operation continues on reception of the BatchReady event.
	return events.ListOf(pbft.eventService.SBEvent(SBCutBatchEvent(pbft.config.MaxBatchSize)))
}

// applyBatchReady processes a new batch ready to be proposed.
// This event is triggered by ISS in response to the CutBatch event produced by this orderer.
func (pbft *pbftInstance) applyBatchReady(batch *isspb.SBBatchReady) *events.EventList {
	eventsOut := events.EmptyList()

	// Clear flag that was set in requestNewBatch(), so that new batches can be requested if necessary.
	pbft.proposal.batchRequested = false

	if pbft.proposal.batchRequestedView == pbft.view {
		// If the protocol is still in the same PBFT view as when the batch was requested, propose the received batch.
		eventsOut.PushBackList(pbft.propose(batch.Batch))
	} else {
		// If the PBFT view advanced since the batch was requested,
		// do not propose the batch and resurrect the requests it contains.
		eventsOut.PushBack(pbft.eventService.SBEvent(SBResurrectBatchEvent(batch.Batch)))
	}

	// Update the number of pending requests that remain after the batch was created.
	eventsOut.PushBackList(pbft.applyPendingRequests(t.NumRequests(batch.PendingRequestsLeft)))

	return eventsOut
}

// propose proposes a new request batch by sending a Preprepare message.
// propose assumes that the state of the PBFT orderer allows sending a new proposal
// and does not perform any checks in this regard.
func (pbft *pbftInstance) propose(batch *requestpb.Batch) *events.EventList {

	// Update proposal counter.
	sn := pbft.segment.SeqNrs[pbft.proposal.proposalsMade]
	pbft.proposal.proposalsMade++

	// Log debug message.
	pbft.logger.Log(logging.LevelDebug, "Proposing.",
		"sn", sn, "batchSize", len(batch.Requests))

	// Create the proposal (consisting of a Preprepare message).
	preprepare := pbftPreprepareMsg(sn, pbft.view, batch, false)

	// Create a Preprepare message send Event.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	msgSendEvent := pbft.eventService.SendMessage(
		PbftPreprepareSBMessage(preprepare),
		pbft.segment.Membership,
	)

	// Set up a new timer for the next proposal.
	timerEvent := pbft.eventService.TimerDelay(
		t.TimeDuration(pbft.config.MaxProposeDelay),
		pbft.eventService.SBEvent(PbftProposeTimeout(uint64(pbft.proposal.proposalsMade+1))),
	)

	// Create a WAL entry and an event to persist it.
	persistEvent := pbft.eventService.WALAppend(PbftPersistPreprepare(preprepare))

	// First the preprepare needs to be persisted to the WAL, and only then it can be sent to the network.
	persistEvent.Next = []*eventpb.Event{msgSendEvent, timerEvent}
	return events.ListOf(persistEvent)
}

// applyMsgPreprepare applies a received preprepare message.
// It performs the necessary checks and, if successful,
// requests a confirmation from ISS that all contained requests have been received and authenticated.
func (pbft *pbftInstance) applyMsgPreprepare(preprepare *isspbftpb.Preprepare, from t.NodeID) *events.EventList {

	// Convenience variable
	sn := t.SeqNr(preprepare.Sn)

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := pbft.preprocessMessage(sn, t.PBFTViewNr(preprepare.View), preprepare, from)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return events.EmptyList()
	}

	// Check that this is the first Preprepare message received.
	// Note that checking the pbft.Preprepared flag instead would be incorrect,
	// as that flag is only set upon receiving the RequestsReady Event.
	if slot.Preprepare != nil {
		pbft.logger.Log(logging.LevelDebug, "Ignoring Preprepare message. Already preprepared or prepreparing.",
			"sn", sn, "from", from)
		return events.EmptyList()
	}

	// Check whether the sender is the legitimate leader of the segment in the current view.
	primary := primaryNode(pbft.segment, pbft.view)
	if from != primary {
		pbft.logger.Log(logging.LevelWarn, "Ignoring Preprepare message. Invalid Leader.",
			"expectedLeader", primary,
			"sender", from,
		)
	}

	// TODO: Check whether the requests belong to the assigned buckets (probably better done at ISS leve).
	// TODO: Check whether the batch contains any duplicate requests.
	//       If it doesn't, mark the contained requests as "in flight" before continuing.

	// Save the received preprepare message.
	slot.Preprepare = preprepare

	// Request the computation of the hash of the Preprepare message.
	return events.ListOf(pbft.eventService.HashRequest(
		[][][]byte{serializePreprepareForHashing(preprepare)},
		preprepareHashOrigin(preprepare)),
	)
}

func (pbft *pbftInstance) applyPreprepareHashResult(digest []byte, preprepare *isspbftpb.Preprepare) *events.EventList {
	eventsOut := events.EmptyList()

	// Convenience variable.
	sn := t.SeqNr(preprepare.Sn)

	// Stop processing the Preprepare if view advanced in the meantime.
	if t.PBFTViewNr(preprepare.View) < pbft.view {
		return events.EmptyList()
	}

	// Save the digest of the Preprepare message and mark the slot as preprepared.
	slot := pbft.slots[pbft.view][sn]
	slot.Digest = digest
	slot.Preprepared = true

	// Send (and persist) a Prepare message.
	eventsOut.PushBackList(pbft.sendPrepare(pbftPrepareMsg(sn, pbft.view, digest)))

	// Advance the state of the pbftSlot even more if necessary
	// (potentially sending a Commit message or even delivering).
	// This is required for the case when the Preprepare message arrives late.
	eventsOut.PushBackList(slot.advanceState(pbft, sn))

	return eventsOut
}

// sendPrepare creates events for persisting and sending a Prepare message.
// The created send event is dependent on (a follow-up event of) the persist event.
func (pbft *pbftInstance) sendPrepare(prepare *isspbftpb.Prepare) *events.EventList {

	// Create persist event.
	persistEvent := pbft.eventService.WALAppend(PbftPersistPrepare(prepare))

	// Append send event as a follow-up.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	persistEvent.FollowUp(pbft.eventService.SendMessage(
		PbftPrepareSBMessage(prepare),
		pbft.segment.Membership,
	))

	// Return a list with a single element - the persist event with the send event as a follow-up.
	return events.ListOf(persistEvent)
}

// applyMsgPrepare applies a received prepare message.
// It performs the necessary checks and, if successful,
// may trigger additional events like the sending of a Commit message.
func (pbft *pbftInstance) applyMsgPrepare(prepare *isspbftpb.Prepare, from t.NodeID) *events.EventList {

	// Convenience variable
	sn := t.SeqNr(prepare.Sn)

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := pbft.preprocessMessage(sn, t.PBFTViewNr(prepare.View), prepare, from)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return events.EmptyList()
	}

	// Check if a Prepare message has already been received from this node in the current view.
	if _, ok := slot.Prepares[from]; ok {
		pbft.logger.Log(logging.LevelDebug, "Ignoring Prepare message. Already received in this view.",
			"sn", sn, "from", from, "view", pbft.view)
		return events.EmptyList()
	}

	// Save the received Prepare message and advance the slot state
	// (potentially sending a Commit message or even delivering).
	slot.Prepares[from] = prepare
	return slot.advanceState(pbft, sn)
}

// sendCommit creates events for persisting and sending a Commit message.
// The created send event is dependent on (a follow-up event of) the persist event.
func (pbft *pbftInstance) sendCommit(commit *isspbftpb.Commit) *events.EventList {

	// Create persist event.
	persistEvent := pbft.eventService.WALAppend(PbftPersistCommit(commit))

	// Append send event as a follow-up.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	persistEvent.FollowUp(pbft.eventService.SendMessage(
		PbftCommitSBMessage(commit),
		pbft.segment.Membership,
	))

	// Return a list with a single element - the persist event with the send event as a follow-up.
	return events.ListOf(persistEvent)
}

// applyMsgCommit applies a received commit message.
// It performs the necessary checks and, if successful,
// may trigger additional events like delivering the corresponding batch.
func (pbft *pbftInstance) applyMsgCommit(commit *isspbftpb.Commit, from t.NodeID) *events.EventList {

	// Convenience variable
	sn := t.SeqNr(commit.Sn)

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := pbft.preprocessMessage(sn, t.PBFTViewNr(commit.View), commit, from)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return events.EmptyList()
	}

	// Check if a Commit message has already been received from this node in the current view.
	if _, ok := slot.Commits[from]; ok {
		pbft.logger.Log(logging.LevelDebug, "Ignoring Commit message. Already received in this view.",
			"sn", sn, "from", from, "view", pbft.view)
		return events.EmptyList()
	}

	// Save the received Commit message and advance the slot state
	// (potentially delivering the corresponding batch).
	slot.Commits[from] = commit
	return slot.advanceState(pbft, sn)
}

// preprocessMessage performs basic checks on a PBFT protocol message based the associated view and sequence number.
// If the message is invalid or outdated, preprocessMessage simply returns nil.
// If the message is from a future view, preprocessMessage will try to buffer it for later processing and returns nil.
// If the message can be processed, preprocessMessage returns the pbftSlot tracking the corresponding state.
// The slot returned by preprocessMessage always belongs to the current view.
func (pbft *pbftInstance) preprocessMessage(sn t.SeqNr, view t.PBFTViewNr, msg proto.Message, from t.NodeID) *pbftSlot {

	if view < pbft.view {
		// Ignore messages from old views.
		pbft.logger.Log(logging.LevelDebug, "Ignoring message from old view.",
			"sn", sn, "from", from, "msgView", view, "localView", pbft.view, "type", fmt.Sprintf("%T", msg))
		return nil
	} else if view > pbft.view || pbft.inViewChange {
		// If message is from a future view, buffer it.
		pbft.logger.Log(logging.LevelDebug, "Buffering message.",
			"sn", sn, "from", from, "msgView", view, "localView", pbft.view, "type", fmt.Sprintf("%T", msg))
		pbft.messageBuffers[from].Store(msg)
		return nil
		// TODO: When view change is implemented, get the messages out of the buffer.
	} else if slot, ok := pbft.slots[pbft.view][sn]; !ok {
		// If the message is from the current view, look up the slot concerned by the message
		// and check if the sequence number is assigned to this PBFT instance.
		pbft.logger.Log(logging.LevelDebug, "Ignoring message. Wrong sequence number.",
			"sn", sn, "from", from, "msgView", view)
		return nil
	} else if slot.Committed {
		// Check if the slot has already been committed (this can happen with state transfer).
		return nil
	} else {
		return slot
	}
}
