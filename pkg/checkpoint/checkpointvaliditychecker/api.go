package checkpointvaliditychecker

import (
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

// The CVC interface represents an implementation of the checkpoint validity checker primitives inside the MirModule module.
// It is responsible for verifying the validity of checkpoints.
type CVC interface {

	// Verify verifies the validity of the checkpoint
	// Returns nil on success (i.e., if the given signature is valid) and a non-nil error otherwise.
	Verify(chkp *checkpointpbtypes.StableCheckpoint, epochNr tt.EpochNr, memberships []*trantorpbtypes.Membership) error
}
