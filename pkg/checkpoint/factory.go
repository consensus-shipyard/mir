package checkpoint

import (
	"time"

	"github.com/filecoin-project/mir/pkg/factorymodule"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
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
			func(submoduleID t.ModuleID, params *factorymodulepb.GeneratorParams) (modules.PassiveModule, error) {

				// Crate a copy of basic module config with an adapted ID for the submodule.
				submc := *mc
				submc.Self = submoduleID

				// Get the instance parameters
				p := params.Type.(*factorymodulepb.GeneratorParams_Checkpoint).Checkpoint
				chkpNodeIDs := t.NodeIDSlice(p.NodeIds)

				protocol := NewProtocol(
					&submc,
					chkpNodeIDs,
					ownID,
					t.SeqNr(p.SeqNr),
					t.EpochNr(p.Epoch),
					t.TimeDuration(time.Duration(p.ResendPeriod)),
					logging.Decorate(logger, "", "chkpSN", p.SeqNr, "chkpEpoch", p.Epoch),
				)

				return protocol, nil
			},
		),
		logger,
	)
}
