package orderers

import (
	"bytes"

	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"

	pbftpbevents "github.com/filecoin-project/mir/pkg/pb/pbftpb/events"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/iss/config"
	ot "github.com/filecoin-project/mir/pkg/orderers/types"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	isspbdsl "github.com/filecoin-project/mir/pkg/pb/isspb/dsl"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// pbftSlot tracks the state of the agreement protocol for one sequence number,
// such as messages received, progress of the reliable broadcast, etc.
type pbftSlot struct {

	// The received preprepare message.
	Preprepare *pbftpbtypes.Preprepare

	// Prepare messages received.
	// A nil entry signifies that an invalid message has been discarded.
	PrepareDigests map[t.NodeID][]byte

	// Valid prepare messages received.
	// Serves mostly as an optimization to not re-validate already validated messages.
	ValidPrepareDigests [][]byte

	// Commit messages received.
	CommitDigests map[t.NodeID][]byte

	// Valid commit messages received.
	// Serves mostly as an optimization to not re-validate already validated messages.
	ValidCommitDigests [][]byte

	// The digest of the proposed (preprepared) certificate
	Digest []byte

	// Flags denoting whether a certificate has been, respectively, preprepared, prepared, and committed in this slot.
	// Note that Preprepared == true is not equivalent to Preprepare != nil, since the Preprepare message is stored
	// before the node preprepares the proposal (it first has to hash the contents of the message
	// and only then can preprepare the proposal and send the Prepare messages).
	Preprepared bool
	Prepared    bool
	Committed   bool

	// Number of nodes executing this instance of PBFT
	numNodes int
}

// newPbftSlot allocates a new pbftSlot object and returns it, initializing all its fields.
// The f parameter designates the number of tolerated failures of the PBFT instance this slot belongs to.
func newPbftSlot(numNodes int) *pbftSlot {
	return &pbftSlot{
		Preprepare:          nil,
		PrepareDigests:      make(map[t.NodeID][]byte),
		ValidPrepareDigests: make([][]byte, 0),
		CommitDigests:       make(map[t.NodeID][]byte),
		ValidCommitDigests:  make([][]byte, 0),
		Digest:              nil,
		Preprepared:         false,
		Prepared:            false,
		Committed:           false,
		numNodes:            numNodes,
	}
}

// populateFromPrevious carries over state from a pbftSlot used in the previous view to this pbftSlot,
// based on the state of the previous slot.
// This is used during view change, when the protocol initializes a new PBFT view.
func (slot *pbftSlot) populateFromPrevious(prevSlot *pbftSlot, view ot.ViewNr) {

	// If the slot has already committed a certificate, just copy over the result.
	if prevSlot.Committed {
		slot.Committed = true
		slot.Digest = prevSlot.Digest
		slot.Preprepare = copyPreprepareToNewView(prevSlot.Preprepare, view)
	}
}

// advanceSlotState checks whether the state of the pbftSlot can be advanced.
// If it can, advanceSlotState updates the state of the pbftSlot and returns a list of Events that result from it.
// Requires the PBFT instance as an argument to use it to generate the proper events.
func (slot *pbftSlot) advanceState(m dsl.Module, pbft *Orderer, sn tt.SeqNr) {
	// If the slot just became prepared, send the Commit message.
	if !slot.Prepared && slot.checkPrepared() {
		slot.Prepared = true

		transportpbdsl.SendMessage(
			m,
			pbft.moduleConfig.Net,
			pbftpbmsgs.Commit(pbft.moduleConfig.Self, sn, pbft.view, slot.Digest),
			pbft.segment.NodeIDs())
	}

	// If the slot just became committed, reset SN timeout and deliver the certificate.
	if !slot.Committed && slot.checkCommitted() {

		// Mark slot as committed.
		slot.Committed = true

		// Set a new SN timeout (unless the segment is finished with a stable checkpoint).
		// Note that we set a new SN timeout even if everything has already been committed.
		// In such a case, we expect a quorum of other nodes to also commit everything and send a Done message
		// before the timeout fires.
		// The timeout event contains the current view and the number of committed slots.
		// It will be ignored if any of those values change by the time the timer fires
		// or if a quorum of nodes confirms having committed all certificates.
		if !pbft.segmentCheckpoint.Stable(len(pbft.segment.Membership.Nodes)) {
			eventpbdsl.TimerDelay(
				m,
				pbft.moduleConfig.Timer,
				[]*eventpbtypes.Event{pbftpbevents.ViewChangeSNTimeout(
					pbft.moduleConfig.Self,
					pbft.view,
					uint64(pbft.numCommitted(pbft.view)))},
				types.Duration(pbft.config.ViewChangeSNTimeout))
		}

		// If all certificates have been committed (i.e. this is the last certificate to commit),
		// send a Done message to all other nodes.
		// This is required for liveness, see comments for pbftSegmentChkp.
		if pbft.allCommitted() {
			pbft.segmentCheckpoint.SetDone()
			pbft.sendDoneMessages(m)
		}

		// Deliver availability certificate (will be verified by ISS)
		isspbdsl.SBDeliver(
			m,
			pbft.moduleConfig.Ord,
			sn,
			slot.Preprepare.Data,
			slot.Preprepare.Aborted,
			pbft.segment.Leader,
			pbft.moduleConfig.Self,
		)
	}
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
	if len(slot.PrepareDigests) < config.StrongQuorum(slot.numNodes) {
		return false
	}

	// Check newly received Prepare messages for validity (whether they contain the hash of the Preprepare message).
	// TODO: Do we need to iterate in a deterministic order here?
	for from, digest := range slot.PrepareDigests {

		// Only check each Prepare message once.
		// When checked, the entry in slot.PrepareDigests is set to nil (but not deleted!)
		// to prevent another Prepare message to be considered again.
		if digest != nil {
			slot.PrepareDigests[from] = nil

			// If the digest in the Prepare message matches that of the Preprepare, add the message to the valid ones.
			if bytes.Equal(digest, slot.Digest) {
				slot.ValidPrepareDigests = append(slot.ValidPrepareDigests, digest)
			}
		}
	}

	// Return true if enough matching Prepare messages have been received.
	return len(slot.ValidPrepareDigests) >= config.StrongQuorum(slot.numNodes)
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
	if len(slot.CommitDigests) < config.StrongQuorum(slot.numNodes) {
		return false
	}

	// Check newly received Commit messages for validity (whether they contain the hash of the Preprepare message).
	// TODO: Do we need to iterate in a deterministic order here?
	for from, digest := range slot.CommitDigests {

		// Only check each Commit message once.
		// When checked, the entry in slot.CommitDigests is set to nil (but not deleted!)
		// to prevent another Commit message to be considered again.
		if digest != nil {
			slot.CommitDigests[from] = nil

			// If the digest in the Commit message matches that of the Preprepare, add the message to the valid ones.
			if bytes.Equal(digest, slot.Digest) {
				slot.ValidCommitDigests = append(slot.ValidCommitDigests, digest)
			}
		}
	}

	// Return true if enough matching Prepare messages have been received.
	return len(slot.ValidCommitDigests) >= config.StrongQuorum(slot.numNodes)
}

func (slot *pbftSlot) getPreprepare(digest []byte) *pbftpbtypes.Preprepare {
	if slot.Preprepared && bytes.Equal(digest, slot.Digest) {
		return slot.Preprepare
	}

	return nil
}

// Note: The Preprepare's view must match the view this slot is assigned to!
func (slot *pbftSlot) catchUp(preprepare *pbftpbtypes.Preprepare, digest []byte) {
	slot.Preprepare = preprepare
	slot.Digest = digest
	slot.Preprepared = true
	slot.Prepared = true
	slot.Committed = true
}
