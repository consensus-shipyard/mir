package multisigcollector

import (
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/parts/batchreconstruction"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/parts/certcreation"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/parts/certverification"

	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/common"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/factorymodule"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig = common.ModuleConfig

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams = common.ModuleParams

// NewModule creates a new instance of the multisig collector module.
// Multisig collector is the simplest implementation of the availability layer.
// Whenever an availability certificate is requested, it pulls a batch from the mempool module,
// sends it to all replicas and collects params.F+1 signatures confirming that
// other nodes have persistently stored the batch.
func NewModule(mc ModuleConfig, params *ModuleParams, logger logging.Logger) (modules.PassiveModule, error) {
	m := dsl.NewModule(mc.Self)

	certcreation.IncludeCreatingCertificates(m, mc, params, logger)
	certverification.IncludeVerificationOfCertificates(m, mc, params)
	batchreconstruction.IncludeBatchReconstruction(m, mc, params, logger)
	return m, nil
}

func NewReconfigurableModule(mc ModuleConfig, logger logging.Logger) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return factorymodule.New(
		mc.Self,
		factorymodule.DefaultParams(

			// This function will be called whenever the factory module
			// is asked to create a new instance of the multisig collector.
			func(mscID t.ModuleID, params *factorypbtypes.GeneratorParams) (modules.PassiveModule, error) {

				// Extract the IDs of the nodes in the membership associated with this instance
				mscParams := params.Type.(*factorypbtypes.GeneratorParams_MultisigCollector).MultisigCollector

				// Create a copy of basic module config with an adapted ID for the submodule.
				submc := mc
				submc.Self = mscID

				// Create a new instance of the multisig collector.
				multisigCollector, err := NewModule(
					submc,
					&ModuleParams{
						// TODO: Use InstanceUIDs properly.
						//       (E.g., concatenate this with the instantiating protocol's InstanceUID when introduced.)
						InstanceUID: []byte(mscID),
						Membership:  mscParams.Membership,
						Limit:       int(mscParams.Limit),
						MaxRequests: int(mscParams.MaxRequests),
					},
					logger,
				)
				if err != nil {
					return nil, err
				}
				return multisigCollector, nil
			},
		),
		logger,
	)
}
