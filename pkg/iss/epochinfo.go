package iss

import (
	lsp "github.com/filecoin-project/mir/pkg/iss/leaderselectionpolicy"
	"github.com/filecoin-project/mir/pkg/orderers/common"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

// epochInfo holds epoch-specific information that becomes irrelevant on advancing to the next epoch.
type epochInfo struct {

	// Epoch number.
	nr tt.EpochNr

	// First sequence number belonging to this epoch.
	firstSN tt.SeqNr

	// This epoch's membership.
	Membership *trantorpbtypes.Membership

	// Orderers' segments associated with the epoch.
	Segments []*common.Segment

	// Leader selection policy.
	leaderPolicy lsp.LeaderSelectionPolicy
}

func newEpochInfo(
	nr tt.EpochNr,
	firstSN tt.SeqNr,
	membership *trantorpbtypes.Membership,
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

func (e *epochInfo) Nr() tt.EpochNr {
	return e.nr
}

func (e *epochInfo) FirstSN() tt.SeqNr {
	return e.firstSN
}

func (e *epochInfo) Len() int {
	l := 0
	for _, segment := range e.Segments {
		l += segment.Len()
	}
	return l
}
