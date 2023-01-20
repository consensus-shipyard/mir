package lookuptxs

import (
	"encoding/hex"
	"github.com/filecoin-project/mir/pkg/dsl"
	mpdsl "github.com/filecoin-project/mir/pkg/mempool/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	mppb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// IncludeTransactionLookupByID registers event handlers for processing RequestTransactions events.
func IncludeTransactionLookupByID(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	commonState *common.State,
) {
	mpdsl.UponRequestTransactions(m, func(txIDs []t.TxID, origin *mppb.RequestTransactionsOrigin) error {
		present := make([]bool, len(txIDs))
		txs := make([]*requestpb.Request, len(txIDs))
		for i, txID := range txIDs {
			txs[i], present[i] = commonState.TxByID[hex.EncodeToString(txID)]
		}

		mpdsl.TransactionsResponse(m, t.ModuleID(origin.Module), present, txs, origin)
		return nil
	})
}
