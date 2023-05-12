package lookuptxs

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/common"
	mpdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mppb "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

// IncludeTransactionLookupByID registers event handlers for processing RequestTransactions events.
func IncludeTransactionLookupByID(
	m dsl.Module,
	_ common.ModuleConfig,
	_ *common.ModuleParams,
	commonState *common.State,
) {
	mpdsl.UponRequestTransactions(m, func(txIDs []tt.TxID, origin *mppb.RequestTransactionsOrigin) error {
		present := make([]bool, len(txIDs))
		txs := make([]*trantorpbtypes.Transaction, len(txIDs))
		for i, txID := range txIDs {
			txs[i], present[i] = commonState.TxByID[txID]
		}

		mpdsl.TransactionsResponse(m, origin.Module, present, txs, origin)
		return nil
	})
}
