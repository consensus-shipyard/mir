package formbatchesext

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	mpdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mppbtypes "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// IncludeBatchCreation registers event handlers for processing NewRequests and RequestBatch events.
func IncludeBatchCreation(
	m dsl.Module,
	mc *common.ModuleConfig,
	fetchTransactions func() []*requestpbtypes.Request,
) {

	mpdsl.UponTransactionIDsResponse(m, func(txIDs []t.TxID, context *requestTxIDsContext) error {
		mpdsl.NewBatch(m, context.origin.Module, txIDs, context.txs, context.origin)
		return nil
	})

	mpdsl.UponRequestBatch(m, func(origin *mppbtypes.RequestBatchOrigin) error {
		txs := fetchTransactions()
		mpdsl.RequestTransactionIDs(m, mc.Self, txs, &requestTxIDsContext{
			txs:    txs,
			origin: origin,
		})
		return nil
	})
}

// Context data structures

type requestTxIDsContext struct {
	txs    []*requestpbtypes.Request
	origin *mppbtypes.RequestBatchOrigin
}
