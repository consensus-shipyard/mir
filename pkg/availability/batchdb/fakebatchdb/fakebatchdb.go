package fakebatchdb

import (
	msctypes "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	batchdbpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/dsl"
	batchdbpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self t.ModuleID // id of this module
}

type moduleState struct {
	batchStore      map[msctypes.BatchID]*batch
	batchesByRetIdx map[tt.RetentionIndex][]msctypes.BatchID
	retIdx          tt.RetentionIndex
}

type batch struct {
	txs []*trantorpbtypes.Transaction
}

// NewModule returns a new module for a fake batch database.
// It stores all the data in memory in plain go maps.
func NewModule(mc ModuleConfig) modules.Module {
	m := dsl.NewModule(mc.Self)

	state := moduleState{
		batchStore:      make(map[msctypes.BatchID]*batch),
		batchesByRetIdx: make(map[tt.RetentionIndex][]msctypes.BatchID),
		retIdx:          0,
	}

	// On StoreBatch request, just store the data in the local memory.
	batchdbpbdsl.UponStoreBatch(m, func(
		batchID msctypes.BatchID,
		txs []*trantorpbtypes.Transaction,
		retIdx tt.RetentionIndex,
		origin *batchdbpbtypes.StoreBatchOrigin,
	) error {
		if retIdx >= state.retIdx {
			b := batch{txs}
			state.batchStore[batchID] = &b
			state.batchesByRetIdx[retIdx] = append(state.batchesByRetIdx[retIdx], batchID)
		}

		batchdbpbdsl.BatchStored(m, origin.Module, origin)
		return nil
	})

	// On LookupBatch request, just check the local map.
	batchdbpbdsl.UponLookupBatch(m, func(batchID msctypes.BatchID, origin *batchdbpbtypes.LookupBatchOrigin) error {

		storedBatch, found := state.batchStore[batchID]
		if !found {
			batchdbpbdsl.LookupBatchResponse(m, origin.Module, false, nil, origin)
			return nil
		}

		batchdbpbdsl.LookupBatchResponse(m, origin.Module, true, storedBatch.txs, origin)
		return nil
	})

	batchdbpbdsl.UponGarbageCollect(m, func(retentionIndex tt.RetentionIndex) error {
		for ; state.retIdx < retentionIndex; state.retIdx++ {
			for _, batchID := range state.batchesByRetIdx[state.retIdx] {
				delete(state.batchStore, batchID)
			}
			delete(state.batchesByRetIdx, state.retIdx)
		}
		return nil
	})

	return m
}
