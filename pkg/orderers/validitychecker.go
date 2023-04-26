package orderers

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/crypto"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
)

// ======================================================================
// permissiveValidityChecker
// (no check performed, everything considered valid)
// ======================================================================

type permissiveValidityChecker struct{}

func newPermissiveValidityChecker() *permissiveValidityChecker {
	return &permissiveValidityChecker{}
}

func (pvc *permissiveValidityChecker) Check(_ []byte) error {
	return nil
}

// ======================================================================
// checkpointValidityChecker
// (for checking checkpoint certificates)
// ======================================================================

type checkpointValidityChecker struct {
	hashImpl     crypto.HashImpl
	certVerifier checkpoint.Verifier
	membership   *trantorpbtypes.Membership
}

func newCheckpointValidityChecker(
	hashImpl crypto.HashImpl,
	certVerifier checkpoint.Verifier,
	membership *trantorpbtypes.Membership,
) *checkpointValidityChecker {
	return &checkpointValidityChecker{
		hashImpl:     hashImpl,
		certVerifier: certVerifier,
		membership:   membership,
	}
}

func (cvc *checkpointValidityChecker) Check(data []byte) error {
	var chkp checkpoint.StableCheckpoint

	if err := chkp.Deserialize(data); err != nil {
		return fmt.Errorf("could not deserialize checkpoint: %w", err)
	}

	if err := chkp.VerifyCert(cvc.hashImpl, cvc.certVerifier, cvc.membership); err != nil {
		return fmt.Errorf("invalid checkpoint certificate: %w", err)
	}

	return nil
}
