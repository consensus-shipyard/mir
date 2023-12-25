package pprepvalidator

import (
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	ppvpbdsl "github.com/filecoin-project/mir/pkg/pb/ordererpb/pprepvalidatorpb/dsl"
	ppvpbtypes "github.com/filecoin-project/mir/pkg/pb/ordererpb/pprepvalidatorpb/types"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/stdmodules/factory"
	t "github.com/filecoin-project/mir/stdtypes"
	"google.golang.org/protobuf/proto"
)

// ModuleConfig sets the module ids.
type ModuleConfig struct {
	Self t.ModuleID
}

// NewModule returns a passive module for the PreprepareValidator module.
func NewModule(mc ModuleConfig, ppv PreprepareValidator) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	ppvpbdsl.UponValidatePreprepare(m, func(preprepare *pbftpbtypes.Preprepare, origin *ppvpbtypes.ValidatePreprepareOrigin) error {
		err := ppv.Check(preprepare)
		ppvpbdsl.PreprepareValidated(m, origin.Module, err, origin)
		return nil
	})

	return m
}

func NewPprepValidatorChkpFactory(mc ModuleConfig,
	hashImpl crypto.HashImpl,
	chkpVerifier checkpoint.Verifier,
	configOffset int,
	logger logging.Logger,
) modules.PassiveModule {

	return factory.New(
		mc.Self,
		factory.DefaultParams(
			func(submoduleID t.ModuleID, params any) (modules.PassiveModule, error) {
				// Crate a copy of basic module config with an adapted ID for the submodule.
				submc := mc
				submc.Self = submoduleID
				// Load parameters from received protobuf
				membership := (*trantorpbtypes.Membership)(params.(*PPVParams))
				// TODO: Use a switch statement and check for a serialized form of the parameters.

				return NewModule(submc, NewCheckpointValidityChecker(hashImpl, chkpVerifier, membership, configOffset, logger)), nil
			},
		),
		logger,
	)

}

func InstanceParams(
	membership *trantorpbtypes.Membership,
) *factorypbtypes.GeneratorParams {
	return &factorypbtypes.GeneratorParams{Type: &factorypbtypes.GeneratorParams_PpvModule{
		PpvModule: &ppvpbtypes.PPrepValidatorChkp{
			Membership: membership,
		},
	}}
}

type PPVParams trantorpbtypes.Membership

func (ppvp *PPVParams) ToBytes() ([]byte, error) {
	return proto.Marshal((*trantorpbtypes.Membership)(ppvp).Pb())
}
