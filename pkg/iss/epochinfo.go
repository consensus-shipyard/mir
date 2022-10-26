package iss

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/pb/isspb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// epochInfo holds epoch-specific information that becomes irrelevant on advancing to the next epoch.
type epochInfo struct {

	// Epoch number.
	nr t.EpochNr

	// First sequence number belonging to this epoch.
	firstSN t.SeqNr

	// The set of IDs of nodes in this epoch's membership.
	nodeIDs map[t.NodeID]struct{}

	// Orderers associated with the epoch.
	Orderers []sbInstance

	// Index of orderers based on the buckets they are assigned.
	// For each bucket ID, this map stores the orderer to which the bucket is assigned in this epoch.
	bucketOrderers map[int]sbInstance

	// TODO: Comment.
	leaderPolicy LeaderSelectionPolicy
}

func newEpochInfo(
	nr t.EpochNr,
	firstSN t.SeqNr,
	nodeIDs []t.NodeID,
	leaderPolicy LeaderSelectionPolicy,
) epochInfo {
	ei := epochInfo{
		nr:             nr,
		firstSN:        firstSN,
		nodeIDs:        membershipSet(nodeIDs),
		Orderers:       nil,
		bucketOrderers: make(map[int]sbInstance),
		leaderPolicy:   leaderPolicy,
	}

	return ei
}

func (e *epochInfo) Nr() t.EpochNr {
	return e.nr
}

func (e *epochInfo) FirstSN() t.SeqNr {
	return e.firstSN
}

func (e *epochInfo) Len() int {
	l := 0
	for _, orderer := range e.Orderers {
		l += len(orderer.Segment().SeqNrs)
	}
	return l
}

// validateSBMessage checks whether an SBMessage is valid in this epoch.
// Returns nil if validation succeeds.
// If validation fails, returns the reason for which the message is considered invalid.
func (e *epochInfo) validateSBMessage(message *isspb.SBMessage, from t.NodeID) error {

	// Message must be destined for this epoch.
	if t.EpochNr(message.Epoch) != e.Nr() {
		return fmt.Errorf("invalid epoch: %v (expected %v)", message.Epoch, e.Nr())
	}

	// Message must refer to a valid SB instance.
	if int(message.Instance) >= len(e.Orderers) {
		return fmt.Errorf("invalid SB instance number: %d", message.Instance)
	}

	// Message must be sent by a node in the current membership.
	if _, ok := e.nodeIDs[from]; !ok {
		return fmt.Errorf("sender of SB message not in the membership: %v", from)
	}

	return nil
}

// membershipSet takes a list of node IDs and returns a map of empty structs with an entry for each node ID in the list.
// The returned map is effectively a set representation of the given list,
// useful for testing whether any given node ID is in the set.
func membershipSet(membership []t.NodeID) map[t.NodeID]struct{} {

	// Allocate a new map representing a set of node IDs
	set := make(map[t.NodeID]struct{})

	// Add an empty struct for each node ID in the list.
	for _, nodeID := range membership {
		set[nodeID] = struct{}{}
	}

	// Return the resulting set of node IDs.
	return set
}
