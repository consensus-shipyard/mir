package formbatches

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	mpdsl "github.com/filecoin-project/mir/pkg/mempool/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	mppb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type State struct {
	*common.State
	NewTxIDs []t.TxID
}

// IncludeBatchCreation registers event handlers for processing new transactions and forming batches.
func IncludeBatchCreation(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	commonState *common.State,
) {
	state := &State{
		State:    commonState,
		NewTxIDs: nil,
	}

	dsl.UponNewRequests(m, func(txs []*requestpb.Request) error {
		mpdsl.RequestTransactionIDs(m, mc.Self, txs, &requestTxIDsContext{txs})
		return nil
	})

	mpdsl.UponTransactionIDsResponse(m, func(txIDs []t.TxID, context *requestTxIDsContext) error {
		for i := range txIDs {
			state.TxByID[txIDs[i]] = context.txs[i]
		}
		state.NewTxIDs = append(state.NewTxIDs, txIDs...)
		return nil
	})

	mpdsl.UponRequestBatch(m, func(origin *mppb.RequestBatchOrigin) error {
		var txIDs []t.TxID
		var txs []*requestpb.Request
		batchSize := 0

		var i int
		var txID t.TxID

		for i, txID = range state.NewTxIDs {
			tx := state.TxByID[txID]

			// TODO: add other limitations (if any) here.
			if i == params.MaxTransactionsInBatch {
				break
			}

			txIDs = append(txIDs, txID)
			txs = append(txs, tx)
			batchSize += len(tx.Data)
		}

		state.NewTxIDs = state.NewTxIDs[i:]

		// Note that a batch may be empty.
		mpdsl.NewBatch(m, t.ModuleID(origin.Module), txIDs, txs, origin)
		return nil
	})
}

// Context data structures

type requestTxIDsContext struct {
	txs []*requestpb.Request
}
