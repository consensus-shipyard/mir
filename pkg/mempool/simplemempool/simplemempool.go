package simplemempool

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/common"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/parts/computeids"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/parts/formbatchesext"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/parts/formbatchesint"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/parts/lookuptxs"
	"github.com/filecoin-project/mir/pkg/modules"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig = common.ModuleConfig

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams = common.ModuleParams

// DefaultModuleConfig returns a valid module config with default names for all modules.
func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		Self:   "availability",
		Hasher: "hasher",
	}
}

func DefaultModuleParams() *ModuleParams {
	return &ModuleParams{
		MaxTransactionsInBatch: 10,
		TxFetcher:              nil,
	}
}

// NewModule creates a new instance of a simple mempool module implementation. It passively waits for
// eventpb.NewRequests events and stores them in a local map.
//
// On a batch request, this implementation creates a batch that consists of as many requests received since the
// previous batch request as possible with respect to params.MaxTransactionsInBatch.
//
// This implementation uses the hash function provided by the mc.Hasher module to compute transaction IDs and batch IDs.
func NewModule(mc *ModuleConfig, params *ModuleParams) modules.Module {
	m := dsl.NewModule(mc.Self)

	commonState := &common.State{
		TxByID: make(map[string]*requestpbtypes.Request),
	}

	computeids.IncludeComputationOfTransactionAndBatchIDs(m, mc, params, commonState)
	lookuptxs.IncludeTransactionLookupByID(m, mc, params, commonState)

	if params.TxFetcher != nil {
		formbatchesext.IncludeBatchCreation(m, mc, params.TxFetcher)
	} else {
		formbatchesint.IncludeBatchCreation(m, mc, params, commonState)
	}

	return m
}
