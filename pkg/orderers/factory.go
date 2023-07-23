package orderers

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/factorymodule"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/orderers/common"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

func Factory(
	mc common.ModuleConfig,
	issParams *issconfig.ModuleParams,
	ownID t.NodeID,
	logger logging.Logger,
) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return factorymodule.New(
		mc.Self,
		factorymodule.DefaultParams(

			// This function will be called whenever the factory module
			// is asked to create a new instance of the Ordering protocol.
			func(submoduleID t.ModuleID, params *factorypbtypes.GeneratorParams) (modules.PassiveModule, error) {

				// Crate a copy of basic module config with an adapted ID for the submodule.
				submc := mc
				submc.Self = submoduleID

				// Load parameters from received protobuf
				p := params.Type.(*factorypbtypes.GeneratorParams_PbftModule).PbftModule
				availabilityID := t.ModuleID(p.AvailabilityId)
				ppvID := t.ModuleID(p.PpvModuleId)
				submc.Ava = availabilityID
				submc.PPrepValidator = ppvID
				epoch := tt.EpochNr(p.Epoch)
				segment := (*common.Segment)(p.Segment)

				// Create new configuration for this particular orderer instance.
				ordererConfig := newOrdererConfig(issParams, segment.NodeIDs(), epoch)

				//TODO better logging here
				logger := logging.Decorate(logger, "", "submoduleID", submoduleID)

				// Select validity checker

				if ppvID == t.ModuleID("pprepvalidatorchkp").Then(t.ModuleID(fmt.Sprintf("%v", epoch))) { // this must change but that's the scope of a different PR
					// TODO: This is a dirty hack! Put (at least the relevant parts of) the configuration in params.
					// Make the agreement on a checkpoint start immediately.
					ordererConfig.MaxProposeDelay = 0
				}

				// Instantiate new protocol instance.
				protocol := NewOrdererModule(
					submc,
					ownID,
					segment,
					ordererConfig,
					logger,
				)

				return protocol, nil
			},
		),
		logger,
	)
}
