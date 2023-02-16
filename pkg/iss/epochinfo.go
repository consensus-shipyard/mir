package iss

import (
	lsp "github.com/filecoin-project/mir/pkg/iss/leaderselectionpolicy"
	"github.com/filecoin-project/mir/pkg/orderers"
	t "github.com/filecoin-project/mir/pkg/types"
)

// epochInfo holds epoch-specific information that becomes irrelevant on advancing to the next epoch.
type epochInfo struct {

	// Epoch number.
	nr t.EpochNr

	// First sequence number belonging to this epoch.
	firstSN t.SeqNr

	// This epoch's membership.
	Membership map[t.NodeID]t.NodeAddress

	// Orderers' segments associated with the epoch.
	Segments []*orderers.Segment

	// TODO: Comment.
	leaderPolicy lsp.LeaderSelectionPolicy
}

func newEpochInfo(
	nr t.EpochNr,
	firstSN t.SeqNr,
	membership map[t.NodeID]t.NodeAddress,
	leaderPolicy lsp.LeaderSelectionPolicy,
) epochInfo {
	ei := epochInfo{
		nr:           nr,
		firstSN:      firstSN,
		Membership:   membership,
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
		l += segment.Len()
	}
	return l
}
