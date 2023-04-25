package formbatchesint

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/common"
	mpdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mppbtypes "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

type State struct {
	*common.State
	NewTxIDs []tt.TxID
}

// IncludeBatchCreation registers event handlers for processing NewRequests and RequestBatch events.
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

	mpdsl.UponNewTransactions(m, func(txs []*trantorpbtypes.Transaction) error {
		mpdsl.RequestTransactionIDs(m, mc.Self, txs, &requestTxIDsContext{txs})
		return nil
	})

	mpdsl.UponTransactionIDsResponse(m, func(txIDs []tt.TxID, context *requestTxIDsContext) error {
		for i := range txIDs {
			if _, ok := state.TxByID[string(txIDs[i])]; !ok {
				state.TxByID[string(txIDs[i])] = context.txs[i]
				state.NewTxIDs = append(state.NewTxIDs, txIDs[i])
			}
		}
		return nil
	})

	mpdsl.UponRequestBatch(m, func(origin *mppbtypes.RequestBatchOrigin) error {
		var txIDs []tt.TxID
		var txs []*trantorpbtypes.Transaction
		batchSize := 0

		txCount := 0
		for _, txID := range state.NewTxIDs {
			tx := state.TxByID[string(txID)]

			// TODO: add other limitations (if any) here.
			if txCount == params.MaxTransactionsInBatch {
				break
			}

			txIDs = append(txIDs, txID)
			txs = append(txs, tx)
			batchSize += len(tx.Data)
			txCount++
		}

		for _, txID := range state.NewTxIDs[:txCount] {
			delete(state.TxByID, string(txID))
		}

		state.NewTxIDs = state.NewTxIDs[txCount:]

		// Note that a batch may be empty.
		mpdsl.NewBatch(m, origin.Module, txIDs, txs, origin)
		return nil
	})
}

// Context data structures

type requestTxIDsContext struct {
	txs []*trantorpbtypes.Transaction
}
