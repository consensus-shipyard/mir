package common

import (
	"fmt"

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
}

func NewCheckpointValidityChecker(
	hashImpl crypto.HashImpl,
	certVerifier checkpoint.Verifier,
	membership *trantorpbtypes.Membership,
) *CheckpointValidityChecker {
	return &CheckpointValidityChecker{
		HashImpl:     hashImpl,
		CertVerifier: certVerifier,
		Membership:   membership,
	}
}

func (cvc *CheckpointValidityChecker) Check(data []byte) error {
	var chkp checkpoint.StableCheckpoint

	if err := chkp.Deserialize(data); err != nil {
		return fmt.Errorf("could not deserialize checkpoint: %w", err)
	}

	if err := chkp.VerifyCert(cvc.HashImpl, cvc.CertVerifier, cvc.Membership); err != nil {
		return fmt.Errorf("invalid checkpoint certificate: %w", err)
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
