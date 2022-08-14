package fakebatchdb

import (
	batchdbdsl "github.com/filecoin-project/mir/pkg/availability/batchdb/dsl"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self t.ModuleID // id of this module
}

// DefaultModuleConfig returns a valid module config with default names for all modules.
func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		Self: "batchdb",
	}
}

type moduleState struct {
	BatchStore       map[t.BatchID]batchInfo
	TransactionStore map[t.TxID]*requestpb.Request
}

type batchInfo struct {
	txIDs    []t.TxID
	metadata []byte
}

// NewModule returns a new module for a fake batch database.
// It stores all the data in memory in plain go maps.
func NewModule(mc *ModuleConfig) modules.Module {
	m := dsl.NewModule(mc.Self)

	state := moduleState{
		BatchStore:       make(map[t.BatchID]batchInfo),
		TransactionStore: make(map[t.TxID]*requestpb.Request),
	}

	// On StoreBatch request, just store the data in the local memory.
	batchdbdsl.UponStoreBatch(m, func(batchID t.BatchID, txIDs []t.TxID, txs []*requestpb.Request, metadata []byte, origin *batchdbpb.StoreBatchOrigin) error {
		state.BatchStore[batchID] = batchInfo{
			txIDs:    txIDs,
			metadata: metadata,
		}

		for i, txID := range txIDs {
			state.TransactionStore[txID] = txs[i]
		}

		batchdbdsl.BatchStored(m, t.ModuleID(origin.Module), origin)
		return nil
	})

	// On LookupBatch request, just check the local map.
	batchdbdsl.UponLookupBatch(m, func(batchID t.BatchID, origin *batchdbpb.LookupBatchOrigin) error {
		info, found := state.BatchStore[batchID]
		if !found {
			batchdbdsl.LookupBatchResponse(m, t.ModuleID(origin.Module), false, nil, nil, origin)
		}

		txs := make([]*requestpb.Request, len(info.txIDs))
		for i, txID := range info.txIDs {
			txs[i] = state.TransactionStore[txID]
		}

		batchdbdsl.LookupBatchResponse(m, t.ModuleID(origin.Module), true, txs, info.metadata, origin)
		return nil
	})

	return m
}
