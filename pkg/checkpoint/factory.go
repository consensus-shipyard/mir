package checkpoint

import (
	"github.com/filecoin-project/mir/pkg/factorymodule"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	"github.com/filecoin-project/mir/stdtypes"
)

func Factory(mc ModuleConfig, ownID stdtypes.NodeID, logger logging.Logger) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return factorymodule.New(
		mc.Self,
		factorymodule.DefaultParams(

			// This function will be called whenever the factory module
			// is asked to create a new instance of the Checkpoint protocol.
			func(submoduleID stdtypes.ModuleID, params *factorypbtypes.GeneratorParams) (modules.PassiveModule, error) {

				// Crate a copy of basic module config with an adapted ID for the submodule.
				submc := mc
				submc.Self = submoduleID

				// Get the instance parameters
				p := params.Type.(*factorypbtypes.GeneratorParams_Checkpoint).Checkpoint

				chkpParams := &ModuleParams{
					OwnID:            ownID,
					Membership:       p.Membership,
					EpochConfig:      p.EpochConfig,
					LeaderPolicyData: p.LeaderPolicyData,
					ResendPeriod:     p.ResendPeriod,
				}

				protocol := NewModule(
					submc,
					chkpParams,
					logging.Decorate(logger, "", "chkpSN", p.EpochConfig.FirstSn, "chkpEpoch", p.EpochConfig.EpochNr),
				)

				return protocol, nil
			},
		),
		logger,
	)
}
