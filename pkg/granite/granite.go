package granite

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/factorymodule"
	"github.com/filecoin-project/mir/pkg/granite/common"
	"github.com/filecoin-project/mir/pkg/granite/internal/parts/consensustask"
	"github.com/filecoin-project/mir/pkg/granite/internal/parts/messagehandlertask"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	granitepbtypes "github.com/filecoin-project/mir/pkg/pb/granitepb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig = common.ModuleConfig

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams = common.ModuleParams

// NewModule creates a new instance of the Granite consensus protocol
func NewModule(mc *ModuleConfig, params *ModuleParams, logger logging.Logger) (modules.PassiveModule, error) {
	m := dsl.NewModule(mc.Self)

	state := &common.State{
		Round:            1,
		Step:             PROPOSE, // First round skip CONVERGE step
		ValidatedMsgs:    NewMsgStore(),
		UnvalidatedMsgs:  NewMsgStore(),
		DecisionMessages: make(map[t.NodeID]map[string]*granitepbtypes.Decision),
	}

	//init data structure to store all messages
	//TODO init maps

	// Include the core logic of the protocol.
	consensustask.IncludeConsensusTask(m, mc, params, state, logger)
	messagehandlertask.IncludeMessageHandlerTask(m, mc, params, state, logger)

	return m, nil
}

func NewReconfigurableModule(mc ModuleConfig, paramsTemplate ModuleParams, logger logging.Logger) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return factorymodule.New(
		mc.Self,
		factorymodule.DefaultParams(

			// This function will be called whenever the factory module
			// is asked to create a new instance of granite.
			func(graniteID t.ModuleID, params *factorypbtypes.GeneratorParams) (modules.PassiveModule, error) {

				// Extract the IDs of the nodes in the membership associated with this instance
				graniteParams := params.Type.(*factorypbtypes.GeneratorParams_Granite).Granite

				// Create a copy of basic module config with an adapted ID for the submodule.
				submc := mc
				submc.Self = graniteID

				// Fill in instance-specific parameters.
				moduleParams := paramsTemplate
				moduleParams.InstanceUID = []byte(graniteID)
				moduleParams.Membership = graniteParams.Membership

				// Create a new instance of granite.
				granite, err := NewModule(
					&submc,
					&moduleParams,
					logger,
				)
				if err != nil {
					return nil, err
				}
				return granite, nil
			},
		),
		logger,
	)
}
