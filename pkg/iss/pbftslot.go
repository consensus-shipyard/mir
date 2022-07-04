package iss

import (
	"bytes"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/isspbftpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// pbftSlot tracks the state of the agreement protocol for one sequence number,
// such as messages received, progress of the reliable broadcast, etc.
type pbftSlot struct {

	// The received preprepare message.
	Preprepare *isspbftpb.Preprepare

	// Prepare messages received.
	// A nil entry signifies that an invalid message has been discarded.
	Prepares map[t.NodeID]*isspbftpb.Prepare

	// Valid prepare messages received.
	// Serves mostly as an optimization to not re-validate already validated messages.
	ValidPrepares []*isspbftpb.Prepare

	// Commit messages received.
	Commits map[t.NodeID]*isspbftpb.Commit

	// Valid commit messages received.
	// Serves mostly as an optimization to not re-validate already validated messages.
	ValidCommits []*isspbftpb.Commit

	// The digest of the proposed (preprepared) batch
	Digest []byte

	// Flags denoting whether a batch has been, respectively, preprepared, prepared, and committed in this slot
	// Note that Preprepared == true is not equivalent to Preprepare != nil, since the Preprepare message is stored
	// before the node preprepares the proposal (it first has to verify that all requests are available
	// and only then can preprepare the proposal and send the Prepare messages).
	Preprepared bool
	Prepared    bool
	Committed   bool

	// Number of tolerated failures of the PBFT instance this slot belongs to.
	f int
}

// newPbftSlot allocates a new pbftSlot object and returns it, initializing all its fields.
// The f parameter designates the number of tolerated failures of the PBFT instance this slot belongs to.
func newPbftSlot(f int) *pbftSlot {
	return &pbftSlot{
		Preprepare:    nil,
		Prepares:      make(map[t.NodeID]*isspbftpb.Prepare),
		ValidPrepares: make([]*isspbftpb.Prepare, 0),
		Commits:       make(map[t.NodeID]*isspbftpb.Commit),
		ValidCommits:  make([]*isspbftpb.Commit, 0),
		Digest:        nil,
		Preprepared:   false,
		Prepared:      false,
		Committed:     false,
		f:             f,
	}
}

// populateFromPrevious carries over state from a pbftSlot used in the previous view to this pbftSlot,
// based on the state of the previous slot.
// This is used during view change, when the protocol initializes a new PBFT view.
func (slot *pbftSlot) populateFromPrevious(prevSlot *pbftSlot, view t.PBFTViewNr) {

	// If the slot has already committed a batch, just copy over the result.
	if prevSlot.Committed {
		slot.Committed = true
		slot.Digest = prevSlot.Digest
		slot.Preprepare = copyPreprepareToNewView(prevSlot.Preprepare, view)
	}
}

// advanceSlotState checks whether the state of the pbftSlot can be advanced.
// If it can, advanceSlotState updates the state of the pbftSlot and returns a list of Events that result from it.
// Requires the PBFT instance as an argument to use it to generate the proper events.
func (slot *pbftSlot) advanceState(pbft *pbftInstance, sn t.SeqNr) *events.EventList {
	eventsOut := events.EmptyList()

	// If the slot just became prepared, send (and persist) the Commit message.
	if !slot.Prepared && slot.checkPrepared() {
		slot.Prepared = true
		eventsOut.PushBackList(pbft.sendCommit(pbftCommitMsg(sn, pbft.view, slot.Digest)))
	}

	// If the slot just became committed, reset batch timeout and deliver the batch.
	if !slot.Committed && slot.checkCommitted() {

		// Mark slot as committed.
		slot.Committed = true

		// Set a new batch timeout (unless the segment is finished with a stable checkpoint).
		// Note that we set a new batch timeout even if everything has already been committed.
		// In such a case, we expect a quorum of other nodes to also commit everything and send a Done message
		// before the timeout fires.
		// The timeout event contains the current view and the number of committed slots.
		// It will be ignored if any of those values change by the time the timer fires
		// or if a quorum of nodes confirms having committed all batches.
		if !pbft.segmentCheckpoint.Stable(len(pbft.segment.Membership)) {
			eventsOut.PushBack(pbft.eventService.TimerDelay(
				t.TimeDuration(pbft.config.ViewChangeBatchTimeout),
				pbft.eventService.SBEvent(PbftViewChangeBatchTimeout(pbft.view, pbft.numCommitted(pbft.view))),
			))
		}

		// If all batches have been committed (i.e. this is the last batch to commit),
		// send a Done message to all other nodes.
		// This is required for liveness, see comments for pbftSegmentChkp.
		if pbft.allCommitted() {
			pbft.segmentCheckpoint.SetDone()
			eventsOut.PushBackList(pbft.sendDoneMessages())
		}

		// Deliver batch.
		eventsOut.PushBack(pbft.eventService.SBEvent(SBDeliverEvent(
			sn,
			slot.Preprepare.Batch,
			slot.Preprepare.Aborted,
		)))

		// TODO: Do we need to persist anything here?
	}

	return eventsOut
}

// checkPrepared evaluates whether the pbftSlot fulfills the conditions to be prepared.
// The slot can be prepared when it has been preprepared and when enough valid Prepare messages have been received.
func (slot *pbftSlot) checkPrepared() bool {

	// The slot must be preprepared first.
	if !slot.Preprepared {
		return false
	}

	// Check if enough unique Prepare messages have been received.
	// (This is just an optimization to allow early returns.)
	if len(slot.Prepares) < 2*slot.f+1 {
		return false
	}

	// Check newly received Prepare messages for validity (whether they contain the hash of the Preprepare message).
	// TODO: Do we need to iterate in a deterministic order here?
	for from, prepare := range slot.Prepares {

		// Only check each Prepare message once.
		// When checked, the entry in slot.Prepares is set to nil (but not deleted!)
		// to prevent another Prepare message to be considered again.
		if prepare != nil {
			slot.Prepares[from] = nil

			// If the digest in the Prepare message matches that of the Preprepare, add the message to the valid ones.
			if bytes.Equal(prepare.Digest, slot.Digest) {
				slot.ValidPrepares = append(slot.ValidPrepares, prepare)
			}
		}
	}

	// Return true if enough matching Prepare messages have been received.
	return len(slot.ValidPrepares) >= 2*slot.f+1
}

// checkCommitted evaluates whether the pbftSlot fulfills the conditions to be committed.
// The slot can be committed when it has been prepared and when enough valid Commit messages have been received.
func (slot *pbftSlot) checkCommitted() bool {

	// The slot must be prepared first.
	if !slot.Prepared {
		return false
	}

	// Check if enough unique Commit messages have been received.
	// (This is just an optimization to allow early returns.)
	if len(slot.Commits) < 2*slot.f+1 {
		return false
	}

	// Check newly received Commit messages for validity (whether they contain the hash of the Preprepare message).
	// TODO: Do we need to iterate in a deterministic order here?
	for from, commit := range slot.Commits {

		// Only check each Commit message once.
		// When checked, the entry in slot.Commits is set to nil (but not deleted!)
		// to prevent another Commit message to be considered again.
		if commit != nil {
			slot.Commits[from] = nil

			// If the digest in the Commit message matches that of the Preprepare, add the message to the valid ones.
			if bytes.Equal(commit.Digest, slot.Digest) {
				slot.ValidCommits = append(slot.ValidCommits, commit)
			}
		}
	}

	// Return true if enough matching Prepare messages have been received.
	return len(slot.ValidCommits) >= 2*slot.f+1
}

func (slot *pbftSlot) getPreprepare(digest []byte) *isspbftpb.Preprepare {
	if slot.Preprepared && bytes.Equal(digest, slot.Digest) {
		return slot.Preprepare
	}

	return nil
}

// Note: The Preprepare's view must match the view this slot is assigned to!
func (slot *pbftSlot) catchUp(preprepare *isspbftpb.Preprepare, digest []byte) {
	slot.Preprepare = preprepare
	slot.Digest = digest
	slot.Preprepared = true
	slot.Prepared = true
	slot.Committed = true
}
