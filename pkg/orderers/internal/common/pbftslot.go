package common

import (
	"bytes"

	ot "github.com/filecoin-project/mir/pkg/orderers/types"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/membutil"
)

// PbftSlot tracks the state of the agreement protocol for one sequence number,
// such as messages received, progress of the reliable broadcast, etc.
type PbftSlot struct {

	// The received preprepare message.
	Preprepare *pbftpbtypes.Preprepare

	// Prepare messages received, indexed by sending node's ID.
	// A nil entry signifies that an invalid message has been discarded.
	PrepareDigests map[t.NodeID][]byte

	// Valid prepare messages received, indexed by sending node's ID.
	// Serves mostly as an optimization to not re-validate already validated messages.
	ValidPrepareDigests map[t.NodeID][]byte

	// Commit messages received, indexed by sending node's ID.
	CommitDigests map[t.NodeID][]byte

	// Valid commit messages received, indexed by sending node's ID.
	// Serves mostly as an optimization to not re-validate already validated messages.
	ValidCommitDigests map[t.NodeID][]byte

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
	membership *trantorpbtypes.Membership
}

// NewPbftSlot allocates a new PbftSlot object and returns it, initializing all its fields.
func NewPbftSlot(membership *trantorpbtypes.Membership) *PbftSlot {
	return &PbftSlot{
		Preprepare:          nil,
		PrepareDigests:      make(map[t.NodeID][]byte),
		ValidPrepareDigests: make(map[t.NodeID][]byte, 0),
		CommitDigests:       make(map[t.NodeID][]byte),
		ValidCommitDigests:  make(map[t.NodeID][]byte, 0),
		Digest:              nil,
		Preprepared:         false,
		Prepared:            false,
		Committed:           false,
		membership:          membership,
	}
}

// PopulateFromPrevious carries over state from a PbftSlot used in the previous view to this PbftSlot,
// based on the state of the previous slot.
// This is used during view change, when the protocol initializes a new PBFT view.
func (slot *PbftSlot) PopulateFromPrevious(prevSlot *PbftSlot, view ot.ViewNr) {

	// If the slot has already committed a certificate, just copy over the result.
	if prevSlot.Committed {
		slot.Committed = true
		slot.Digest = prevSlot.Digest
		preprepareCopy := *prevSlot.Preprepare
		preprepareCopy.View = view
		slot.Preprepare = &preprepareCopy
	}
}

// CheckPrepared evaluates whether the PbftSlot fulfills the conditions to be prepared.
// The slot can be prepared when it has been preprepared and when enough valid Prepare messages have been received.
func (slot *PbftSlot) CheckPrepared() bool {

	// The slot must be preprepared first.
	if !slot.Preprepared {
		return false
	}

	// Check if enough unique Prepare messages have been received.
	// (This is just an optimization to allow early returns.)
	if !membutil.HaveStrongQuorum(slot.membership, maputil.GetKeys(slot.PrepareDigests)) {
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
				slot.ValidPrepareDigests[from] = digest
			}
		}
	}

	// Return true if enough matching Prepare messages have been received.
	return membutil.HaveStrongQuorum(slot.membership, maputil.GetKeys(slot.ValidPrepareDigests))
}

// CheckCommitted evaluates whether the PbftSlot fulfills the conditions to be committed.
// The slot can be committed when it has been prepared and when enough valid Commit messages have been received.
func (slot *PbftSlot) CheckCommitted() bool {

	// The slot must be prepared first.
	if !slot.Prepared {
		return false
	}

	// Check if enough unique Commit messages have been received.
	// (This is just an optimization to allow early returns.)
	if !membutil.HaveStrongQuorum(slot.membership, maputil.GetKeys(slot.CommitDigests)) {
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
				slot.ValidCommitDigests[from] = digest
			}
		}
	}

	// Return true if enough matching Prepare messages have been received.
	return membutil.HaveStrongQuorum(slot.membership, maputil.GetKeys(slot.ValidCommitDigests))
}

func (slot *PbftSlot) GetPreprepare(digest []byte) *pbftpbtypes.Preprepare {
	if slot.Preprepared && bytes.Equal(digest, slot.Digest) {
		return slot.Preprepare
	}

	return nil
}

// Note: The Preprepare's view must match the view this slot is assigned to!
func (slot *PbftSlot) CatchUp(preprepare *pbftpbtypes.Preprepare, digest []byte) {
	slot.Preprepare = preprepare
	slot.Digest = digest
	slot.Preprepared = true
	slot.Prepared = true
	slot.Committed = true
}
