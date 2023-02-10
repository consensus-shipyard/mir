package fakebatchdb

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/emptybatchid"
	"github.com/filecoin-project/mir/pkg/modules"
	batchdbpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/dsl"
	batchdbpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
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

type txIDString string

type moduleState struct {
	BatchStore       map[t.BatchIDString]batchInfo
	TransactionStore map[txIDString]*requestpbtypes.Request
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
		BatchStore:       make(map[t.BatchIDString]batchInfo),
		TransactionStore: make(map[txIDString]*requestpbtypes.Request),
	}

	// On StoreBatch request, just store the data in the local memory.
	batchdbpbdsl.UponStoreBatch(m, func(batchID t.BatchID, txIDs []t.TxID, txs []*requestpbtypes.Request, metadata []byte, origin *batchdbpbtypes.StoreBatchOrigin) error {
		state.BatchStore[t.BatchIDString(batchID)] = batchInfo{
			txIDs:    txIDs,
			metadata: metadata,
		}

		for i, txID := range txIDs {
			state.TransactionStore[txIDString(txID)] = txs[i]
		}

		batchdbpbdsl.BatchStored(m, origin.Module, origin)
		return nil
	})

	// On LookupBatch request, just check the local map.
	batchdbpbdsl.UponLookupBatch(m, func(batchID t.BatchID, origin *batchdbpbtypes.LookupBatchOrigin) error {
		if emptybatchid.IsEmptyBatchID(batchID) {
			batchdbpbdsl.LookupBatchResponse(m, origin.Module, true, []*requestpbtypes.Request{}, []uint8{}, origin)
		}

		info, found := state.BatchStore[t.BatchIDString(batchID)]
		if !found {
			batchdbpbdsl.LookupBatchResponse(m, origin.Module, false, nil, nil, origin)
			return nil
		}

		txs := make([]*requestpbtypes.Request, len(info.txIDs))
		for i, txID := range info.txIDs {
			txs[i] = state.TransactionStore[txIDString(txID)]
		}

		batchdbpbdsl.LookupBatchResponse(m, origin.Module, true, txs, info.metadata, origin)
		return nil
	})

	return m
}
