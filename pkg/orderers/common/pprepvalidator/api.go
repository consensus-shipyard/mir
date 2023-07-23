package pprepvalidator

import pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"

// PreprepareValidator is the interface of an external checker of validity of proposed data.
type PreprepareValidator interface {

	// Check verifies the validity of the checkpoint
	// Returns nil on success (i.e., if the given signature is valid) and a non-nil error otherwise.
	Check(preprepare *pbftpbtypes.Preprepare) error
}
