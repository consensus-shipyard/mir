package chkpvalidator

import (
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

// The ChkpValidator interface represents an implementation of
// the checkpoint validity checker primitives inside the MirModule module.
// It is responsible for verifying the validity of checkpoints.
type ChkpValidator interface {

	// Verify verifies the validity of the checkpoint
	// Returns nil on success (i.e., if the given checkpoint is valid) and a non-nil error otherwise.
	// The epoch and memberships parameters must be given because the checkpoint validity checker must be able
	// to retrieve the membership for the given epoch in order to verify the signatures of the strong quorum
	// that certifies the stable the checkpoint
	Verify(chkp *checkpointpbtypes.StableCheckpoint, epochNr tt.EpochNr, memberships []*trantorpbtypes.Membership) error
}
