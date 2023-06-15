package common

import (
	es "github.com/go-errors/errors"

	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
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

func (pvc *PermissiveValidityChecker) Check(_ []byte) error {
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
	issParams    *issconfig.ModuleParams
	logger       logging.Logger
}

func NewCheckpointValidityChecker(
	hashImpl crypto.HashImpl,
	certVerifier checkpoint.Verifier,
	membership *trantorpbtypes.Membership,
	issParams *issconfig.ModuleParams,
	logger logging.Logger,
) *CheckpointValidityChecker {
	return &CheckpointValidityChecker{
		HashImpl:     hashImpl,
		CertVerifier: certVerifier,
		Membership:   membership,
		issParams:    issParams,
		logger:       logger,
	}
}

func (cvc *CheckpointValidityChecker) Check(data []byte) error {
	var chkp checkpoint.StableCheckpoint

	if err := chkp.Deserialize(data); err != nil {
		return es.Errorf("could not deserialize checkpoint: %w", err)
	}

	if err := chkp.Verify(cvc.issParams, cvc.HashImpl, cvc.CertVerifier, cvc.Membership, cvc.logger); err != nil {
		return es.Errorf("invalid checkpoint: %w", err)
	}

	return nil
}

// ValidityChecker is the interface of an external checker of validity of proposed data.
// Each orderer is provided with an object implementing this interface
// and applies its Check method to all received proposals.
type ValidityChecker interface {

	// Check returns nil if the provided proposal data is valid, a non-nil error otherwise.
	Check(data []byte) error
}
