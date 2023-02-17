package formbatches

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	mpdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mppbtypes "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type State struct {
	*common.State
	NewTxIDs []t.TxID
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

	eventpbdsl.UponNewRequests(m, func(txs []*requestpbtypes.Request) error {
		mpdsl.RequestTransactionIDs(m, mc.Self, txs, &requestTxIDsContext{txs})
		return nil
	})

	mpdsl.UponTransactionIDsResponse(m, func(txIDs []t.TxID, context *requestTxIDsContext) error {
		for i := range txIDs {
			state.TxByID[string(txIDs[i])] = context.txs[i]
		}
		state.NewTxIDs = append(state.NewTxIDs, txIDs...)
		return nil
	})

	mpdsl.UponRequestBatch(m, func(origin *mppbtypes.RequestBatchOrigin) error {
		var txIDs []t.TxID
		var txs []*requestpbtypes.Request
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

		state.NewTxIDs = state.NewTxIDs[txCount:]

		// Note that a batch may be empty.
		mpdsl.NewBatch(m, origin.Module, txIDs, txs, origin)
		return nil
	})
}

// Context data structures

type requestTxIDsContext struct {
	txs []*requestpbtypes.Request
}
