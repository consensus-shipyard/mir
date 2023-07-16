package prepreparevaliditychecker

import (
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/factorymodule"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	pvcpbdsl "github.com/filecoin-project/mir/pkg/pb/ordererpb/prepreparevaliditycheckerpb/dsl"
	pvcpbtypes "github.com/filecoin-project/mir/pkg/pb/ordererpb/prepreparevaliditycheckerpb/types"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids.
type ModuleConfig struct {
	Self t.ModuleID
	CVC  t.ModuleID
}

// NewModule returns a passive module for the PreprepareValidityChecker module.
func NewModule(mc ModuleConfig, pvc PreprepareValidityChecker) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	pvcpbdsl.UponValidatepreprepare(m, func(preprepare *pbftpbtypes.Preprepare, origin *pvcpbtypes.ValidatePreprepareOrigin) error {
		err := pvc.Check(preprepare)
		pvcpbdsl.PreprepareValidated(m, origin.Module, err, origin)
		return nil
	})

	return m
}

func NewFactory(mc ModuleConfig,
	hashImpl crypto.HashImpl,
	chkpVerifier checkpoint.Verifier,
	issParams *issconfig.ModuleParams,
	logger logging.Logger,
) modules.PassiveModule {

	return factorymodule.New(
		mc.Self,
		factorymodule.DefaultParams(
			func(submoduleID t.ModuleID, params *factorypbtypes.GeneratorParams) (modules.PassiveModule, error) {
				// Crate a copy of basic module config with an adapted ID for the submodule.
				submc := mc
				submc.Self = submoduleID
				// Load parameters from received protobuf
				p := params.Type.(*factorypbtypes.GeneratorParams_PvcModule).PvcModule
				// Select validity checker
				var validityChecker PreprepareValidityChecker
				switch ValidityCheckerType(p.ValidityChecker) {
				case PermissiveVC:
					validityChecker = NewPermissiveValidityChecker()
				case CheckpointVC:
					validityChecker = NewCheckpointValidityChecker(hashImpl, chkpVerifier, p.Membership, issParams, logger)
				}
				return NewModule(submc, validityChecker), nil
			},
		),
		logger,
	)

}

func InstanceParams(
	validityChecker ValidityCheckerType,
	membership *trantorpbtypes.Membership,
) *factorypbtypes.GeneratorParams {
	return &factorypbtypes.GeneratorParams{Type: &factorypbtypes.GeneratorParams_PvcModule{
		PvcModule: &pvcpbtypes.PVCModule{
			ValidityChecker: uint64(validityChecker),
			Membership:      membership,
		},
	}}
}
