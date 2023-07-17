package common

import (
	es "github.com/go-errors/errors"

	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"

	"github.com/filecoin-project/mir/pkg/logging"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/crypto"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
)

// ======================================================================
// PermissiveValidityChecker
// (no check performed, everything considered valid)
// ======================================================================

type PermissiveValidityChecker struct{}

func NewPermissiveValidityChecker() *PermissiveValidityChecker {
	return &PermissiveValidityChecker{}
}

func (pvc *PermissiveValidityChecker) Check(preprepare *pbftpbtypes.Preprepare) error {
	if preprepare.Data == nil && !preprepare.Aborted {
		return es.Errorf("invalid preprepare: data is nil")
	}
	return nil
}

// ======================================================================
// CheckpointValidityChecker
// (for checking checkpoint certificates)
// ======================================================================

type CheckpointValidityChecker struct {
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
) *CheckpointValidityChecker {
	return &CheckpointValidityChecker{
		HashImpl:     hashImpl,
		CertVerifier: certVerifier,
		Membership:   membership,
		configOffset: configOffset,
		logger:       logger,
	}
}

func (cv *CheckpointValidityChecker) Check(preprepare *pbftpbtypes.Preprepare) error {
	var chkp checkpoint.StableCheckpoint

	if err := chkp.Deserialize(preprepare.Data); err != nil {
		return es.Errorf("could not deserialize checkpoint: %w", err)
	}

	if err := chkp.Verify(cv.configOffset, cv.HashImpl, cv.CertVerifier, cv.Membership); err != nil {
		return es.Errorf("invalid checkpoint: %w", err)
	}

	return nil
}

// ValidityChecker is the interface of an external checker of validity of proposed data.
// Each orderer is provided with an object implementing this interface
// and applies its Check method to all received proposals.
type ValidityChecker interface {

	// Check returns nil if the provided proposal data is valid, a non-nil error otherwise.
	Check(preprepare *pbftpbtypes.Preprepare) error
}
