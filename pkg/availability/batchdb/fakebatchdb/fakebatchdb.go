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

	// batchStore stores the actual batch data.
	batchStore map[msctypes.BatchID]*batch

	// batchesByRetIdx is an index that stores a set of batches associated with each particular retention index
	// that has not yet been garbage-collected.
	// Upon garbage-collection of a particular retention index, all the batches stored here under that retention index
	// are deleted, unless they are also associated with a higher retention index.
	// Using a set of batches rather than a list prevents duplicate entries.
	batchesByRetIdx map[tt.RetentionIndex]map[msctypes.BatchID]struct{}

	// retIdx is the lowest retention index that has not yet been garbage-collected.
	retIdx tt.RetentionIndex
}

type batch struct {

	// Transactions in the batch.
	txs []*trantorpbtypes.Transaction

	// The maximal retention index with which the batch was stored.
	// Note that the same batch can be stored in the batchDB (under the same batch ID)
	// multiple times if, e.g., multiple nodes propose the same transactions in different epoch in Trantor.
	// Thus, when garbage-collecting, we must not delete a batch that still needs to be retained
	// for a higher retention index.
	maxRetIdx tt.RetentionIndex
}

// NewModule returns a new module for a fake batch database.
// It stores all the data in memory in plain go maps.
func NewModule(mc ModuleConfig) modules.Module {
	m := dsl.NewModule(mc.Self)

	state := moduleState{
		batchStore:      make(map[msctypes.BatchID]*batch),
		batchesByRetIdx: make(map[tt.RetentionIndex]map[msctypes.BatchID]struct{}),
		retIdx:          0,
	}

	// On StoreBatch request, just store the data in the local memory.
	batchdbpbdsl.UponStoreBatch(m, func(
		batchID msctypes.BatchID,
		txs []*trantorpbtypes.Transaction,
		retIdx tt.RetentionIndex,
		origin *batchdbpbtypes.StoreBatchOrigin,
	) error {

		// Only save the batch if its retention index has not yet been garbage-collected.
		if retIdx >= state.retIdx {
			// Check if we already have the batch.
			b, ok := state.batchStore[batchID]
			if !ok || b.maxRetIdx < retIdx {
				// If we do not, or if the stored batch's retention index is lower,
				// store the received batch with the up-to-date retention index
				state.batchStore[batchID] = &batch{txs, retIdx}
			}

			if _, ok := state.batchesByRetIdx[retIdx]; !ok {
				state.batchesByRetIdx[retIdx] = make(map[msctypes.BatchID]struct{})
			}
			state.batchesByRetIdx[retIdx][batchID] = struct{}{}
		}

		// Note that we emit a BatchStored event even if the batch's retention index was too low
		// (and thus the batch was not actually stored).
		// However, since this situation is indistinguishable from
		// storing the batch and immediately garbage-collecting it,
		// it is simpler to report success to the module that produced the StoreBatch event
		// (rather than creating a whole different code branch with no real utility).
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
			for batchID := range state.batchesByRetIdx[state.retIdx] {
				if state.batchStore[batchID].maxRetIdx <= state.retIdx {
					delete(state.batchStore, batchID)
				}
			}
			delete(state.batchesByRetIdx, state.retIdx)
		}
		return nil
	})

	return m
}
