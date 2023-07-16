package prepreparevaliditychecker

import pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"

// PreprepareValidityChecker is the interface of an external checker of validity of proposed data.
type PreprepareValidityChecker interface {

	// Verify verifies the validity of the checkpoint
	// Returns nil on success (i.e., if the given signature is valid) and a non-nil error otherwise.
	Check(preprepare *pbftpbtypes.Preprepare) error
}

type ValidityCheckerType uint64

const (
	PermissiveVC = iota
	CheckpointVC
)
