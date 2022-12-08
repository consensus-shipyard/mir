package iss

import (
	"github.com/filecoin-project/mir/pkg/orderers"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/issutil"
)

// epochInfo holds epoch-specific information that becomes irrelevant on advancing to the next epoch.
type epochInfo struct {

	// Epoch number.
	nr t.EpochNr

	// First sequence number belonging to this epoch.
	firstSN t.SeqNr

	// The set of IDs of nodes in this epoch's membership.
	nodeIDs map[t.NodeID]struct{}

	// Orderers' segments associated with the epoch.
	Segments []*orderers.Segment

	// TODO: Comment.
	leaderPolicy issutil.LeaderSelectionPolicy
}

func newEpochInfo(
	nr t.EpochNr,
	firstSN t.SeqNr,
	nodeIDs []t.NodeID,
	leaderPolicy issutil.LeaderSelectionPolicy,
) epochInfo {
	ei := epochInfo{
		nr:           nr,
		firstSN:      firstSN,
		nodeIDs:      membershipSet(nodeIDs),
		leaderPolicy: leaderPolicy,
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
	for _, segment := range e.Segments {
		l += len(segment.SeqNrs)
	}
	return l
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
