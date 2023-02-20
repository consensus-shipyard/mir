package orderers

import (
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/factorymodule"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
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
			func(submoduleID t.ModuleID, params *factorymodulepb.GeneratorParams) (modules.PassiveModule, error) {

				// Crate a copy of basic module config with an adapted ID for the submodule.
				submc := *mc
				submc.Self = submoduleID

				p := params.Type.(*factorymodulepb.GeneratorParams_PbftModule).PbftModule

				availabilityID := t.ModuleID(p.AvailabilityId)
				submc.Ava = availabilityID

				epoch := t.EpochNr(p.Epoch)
				// Get the segment parameters
				seqNrs := t.SeqNrSlice(p.Segment.SeqNrs)
				leader := t.NodeID(p.Segment.Leader)

				segment := &Segment{
					Leader:     leader,
					Membership: t.Membership(p.Segment.Membership),
					SeqNrs:     seqNrs,
				}

				// Select validity checker
				var validityChecker ValidityChecker
				switch ValidityCheckerType(p.ValidityChecker) {
				case PermissiveValidityChecker:
					validityChecker = newPermissiveValidityChecker()
				case CheckpointValidityChecker:
					validityChecker = newCheckpointValidityChecker(hashImpl, chkpVerifier, segment.Membership)
				}

				protocol := NewOrdererModule(
					&submc,
					ownID,
					segment,
					newOrdererConfig(
						issParams,
						segment.NodeIDs(),
						epoch,
					),
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
