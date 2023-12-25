package multisigcollector

import (
	"math"

	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/common"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/parts/batchreconstruction"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/parts/certcreation"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/parts/certverification"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	"github.com/filecoin-project/mir/stdmodules/factory"
	t "github.com/filecoin-project/mir/stdtypes"
	"google.golang.org/protobuf/proto"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig = common.ModuleConfig

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams = common.ModuleParams

// DefaultParamsTemplate returns the availability module parameters structure partially filled with default values.
// Fields without a meaningful default value (like InstanceUID, Epoch, and Membership)
// are left empty (zero values for their corresponding type).
func DefaultParamsTemplate() ModuleParams {
	return ModuleParams{
		Limit: 1, // Number of sub-certificates in one availability certificate.
		// TODO: Increase the Limit when https://github.com/filecoin-project/mir/issues/495 is fixed.
		MaxRequests: math.MaxInt, // By default, have the availability module run (basically) forever.
	}
}

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

func NewReconfigurableModule(mc ModuleConfig, paramsTemplate ModuleParams, logger logging.Logger) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return factory.New(
		mc.Self,
		factory.DefaultParams(

			// This function will be called whenever the factory module
			// is asked to create a new instance of the multisig collector.
			func(mscID t.ModuleID, params any) (modules.PassiveModule, error) {

				// Extract the IDs of the nodes in the membership associated with this instance
				// TODO: Use a switch statement and check for a serialized form of the parameters.
				mscParams := (*mscpbtypes.InstanceParams)(params.(*InstanceParams))

				// Create a copy of basic module config with an adapted ID for the submodule.
				submc := mc
				submc.Self = mscID

				// Fill in instance-specific parameters.
				moduleParams := paramsTemplate
				moduleParams.InstanceUID = []byte(mscID)
				moduleParams.EpochNr = mscParams.Epoch
				moduleParams.Membership = mscParams.Membership
				moduleParams.MaxRequests = int(mscParams.MaxRequests)
				// TODO: Use InstanceUIDs properly.
				//       (E.g., concatenate this with the instantiating protocol's InstanceUID when introduced.)

				// Create a new instance of the multisig collector.
				multisigCollector, err := NewModule(
					submc,
					&moduleParams,
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

type InstanceParams mscpbtypes.InstanceParams

func (ip *InstanceParams) ToBytes() ([]byte, error) {
	return proto.Marshal((*mscpbtypes.InstanceParams)(ip).Pb())
}
