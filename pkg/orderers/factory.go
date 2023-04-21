package orderers

import (
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/factorymodule"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type ValidityCheckerType uint64

const (
	PermissiveValidityChecker = iota
	CheckpointValidityChecker
)

func Factory(
	mc *ModuleConfig,
	issParams *issconfig.ModuleParams,
	ownID t.NodeID,
	hashImpl crypto.HashImpl,
	chkpVerifier checkpoint.Verifier,
	logger logging.Logger,
) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return factorymodule.New(
		mc.Self,
		factorymodule.DefaultParams(

			// This function will be called whenever the factory module
			// is asked to create a new instance of the Checkpoint protocol.
			func(submoduleID t.ModuleID, params *factorypbtypes.GeneratorParams) (modules.PassiveModule, error) {

				// Crate a copy of basic module config with an adapted ID for the submodule.
				submc := *mc
				submc.Self = submoduleID

				// Load parameters from received protobuf
				p := params.Type.(*factorypbtypes.GeneratorParams_PbftModule).PbftModule
				availabilityID := t.ModuleID(p.AvailabilityId)
				submc.Ava = availabilityID
				epoch := tt.EpochNr(p.Epoch)
				segment := (*Segment)(p.Segment)

				// Create new configuration for this particular orderer instance.
				ordererConfig := newOrdererConfig(issParams, segment.NodeIDs(), epoch)

				// Select validity checker
				var validityChecker ValidityChecker
				switch ValidityCheckerType(p.ValidityChecker) {
				case PermissiveValidityChecker:
					validityChecker = newPermissiveValidityChecker()
				case CheckpointValidityChecker:
					validityChecker = newCheckpointValidityChecker(hashImpl, chkpVerifier, segment.Membership)

					// TODO: This is a dirty hack! Put (at least the relevant parts of) the configuration in params.
					// Make the agreement on a checkpoint start immediately.
					ordererConfig.MaxProposeDelay = 0

				}

				// Instantiate new protocol instance.
				protocol := NewOrdererModule(
					&submc,
					ownID,
					segment,
					ordererConfig,
					validityChecker,

					//TODO better logging here
					logging.Decorate(logger, "", "submoduleID", submoduleID),
				)

				return protocol, nil
			},
		),
		logger,
	)
}
