package checkpoint

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/stdmodules/factory"
	"github.com/filecoin-project/mir/stdtypes"
)

func Factory(mc ModuleConfig, logger logging.Logger) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return factory.New(
		mc.Self,
		factory.DefaultParams(

			// This function will be called whenever the factory module
			// is asked to create a new instance of the Checkpoint protocol.
			func(submoduleID stdtypes.ModuleID, params any) (modules.PassiveModule, error) {

				// Crate a copy of basic module config with an adapted ID for the submodule.
				submc := mc
				submc.Self = submoduleID

				// Get the instance parameters
				chkpParams := params.(*ModuleParams)

				protocol := NewModule(
					submc,
					chkpParams,
					logging.Decorate(logger, "",
						"chkpSN", chkpParams.EpochConfig.FirstSn,
						"chkpEpoch", chkpParams.EpochConfig.EpochNr),
				)

				return protocol, nil
			},
		),
		logger,
	)
}
