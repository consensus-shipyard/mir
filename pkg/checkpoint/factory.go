package checkpoint

import (
	"time"

	"github.com/filecoin-project/mir/pkg/factorymodule"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/factorypb"
	"github.com/filecoin-project/mir/pkg/timer/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

func Factory(mc *ModuleConfig, ownID t.NodeID, logger logging.Logger) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return factorymodule.New(
		mc.Self,
		factorymodule.DefaultParams(

			// This function will be called whenever the factory module
			// is asked to create a new instance of the Checkpoint protocol.
			func(submoduleID t.ModuleID, params *factorypb.GeneratorParams) (modules.PassiveModule, error) {

				// Crate a copy of basic module config with an adapted ID for the submodule.
				submc := *mc
				submc.Self = submoduleID

				// Get the instance parameters
				p := params.Type.(*factorypb.GeneratorParams_Checkpoint).Checkpoint

				protocol := NewProtocol(
					&submc,
					ownID,
					t.Membership(p.Membership),
					p.EpochConfig,
					p.LeaderPolicyData,
					types.Duration(time.Duration(p.ResendPeriod)),
					logging.Decorate(logger, "", "chkpSN", p.EpochConfig.FirstSn, "chkpEpoch", p.EpochConfig.EpochNr),
				)

				return protocol, nil
			},
		),
		logger,
	)
}
