package orderers

import (
	"strconv"

	"github.com/filecoin-project/mir/pkg/factorymodule"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/issutil"
)

func Factory(mc *ModuleConfig, issParams *issutil.ModuleParams, ownID t.NodeID, logger logging.Logger) modules.PassiveModule {
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

				p := params.Type.(*factorymodulepb.GeneratorParams_PbftModule).PbftModule

				availabilityID := t.ModuleID(p.AvailabilityId)
				submc.Ava = availabilityID
				epoch, err := strconv.ParseUint(string(submc.Self.Sub()[0:1]), 10, 64)
				if err != nil {
					return nil, err
				}

				epochNr := t.EpochNr(epoch)
				// Get the segment parameters
				membership := t.NodeIDSlice(p.Segment.Membership)
				seqNrs := t.SeqNrSlice(p.Segment.SeqNrs)
				leader := t.NodeID(p.Segment.Leader)

				segment := &Segment{
					Leader:     leader,
					Membership: membership,
					SeqNrs:     seqNrs,
				}

				protocol := NewOrdererModule(
					&submc,
					ownID,
					segment,
					newOrdererConfig(
						issParams,
						membership,
						epochNr,
					),

					//TODO better logging here
					logging.Decorate(logger, "", "submoduleID", submoduleID),
				)

				return protocol, nil
			},
		),
		logger,
	)
}
