package pprepvalidator

import (
	es "github.com/go-errors/errors"

	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"

	"github.com/filecoin-project/mir/pkg/logging"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/crypto"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
)

// ======================================================================
// PermissivePreprepareValidator
// (no check performed, everything considered valid)
// ======================================================================

type PermissivePreprepareValidator struct{}

func NewPermissiveValidityChecker() *PermissivePreprepareValidator {
	return &PermissivePreprepareValidator{}
}

func (ppv *PermissivePreprepareValidator) Check(preprepare *pbftpbtypes.Preprepare) error {
	if preprepare.Data == nil && !preprepare.Aborted {
		return es.Errorf("invalid preprepare: data is nil")
	}
	return nil
}

// ======================================================================
// CheckpointPreprepareValidator
// (for checking checkpoint certificates)
// ======================================================================

type CheckpointPreprepareValidator struct {
	HashImpl     crypto.HashImpl
	CertVerifier checkpoint.Verifier
	Membership   *trantorpbtypes.Membership
	configOffset int
	logger       logging.Logger
}

func NewCheckpointValidityChecker(
	hashImpl crypto.HashImpl,
	certVerifier checkpoint.Verifier,
	membership *trantorpbtypes.Membership,
	configOffset int,
	logger logging.Logger,
) *CheckpointPreprepareValidator {
	return &CheckpointPreprepareValidator{
		HashImpl:     hashImpl,
		CertVerifier: certVerifier,
		Membership:   membership,
		configOffset: configOffset,
		logger:       logger,
	}
}

func (cv *CheckpointPreprepareValidator) Check(preprepare *pbftpbtypes.Preprepare) error {
	var chkp checkpoint.StableCheckpoint

	if err := chkp.Deserialize(preprepare.Data); err != nil {
		return es.Errorf("could not deserialize checkpoint: %w", err)
	}

	if err := chkp.Verify(cv.configOffset, cv.HashImpl, cv.CertVerifier, cv.Membership); err != nil {
		return es.Errorf("invalid checkpoint: %w", err)
	}

	return nil
}
