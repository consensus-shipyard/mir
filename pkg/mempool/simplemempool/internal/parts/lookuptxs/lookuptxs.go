package lookuptxs

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	mpdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mppb "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
)

// IncludeTransactionLookupByID registers event handlers for processing RequestTransactions events.
func IncludeTransactionLookupByID(
	m dsl.Module,
	_ *common.ModuleConfig,
	_ *common.ModuleParams,
	commonState *common.State,
) {
	mpdsl.UponRequestTransactions(m, func(txIDs [][]uint8, origin *mppb.RequestTransactionsOrigin) error {
		present := make([]bool, len(txIDs))
		txs := make([]*requestpbtypes.Request, len(txIDs))
		for i, txID := range txIDs {
			txs[i], present[i] = commonState.TxByID[string(txID)]
		}

		mpdsl.TransactionsResponse(m, origin.Module, present, txs, origin)
		return nil
	})
}
