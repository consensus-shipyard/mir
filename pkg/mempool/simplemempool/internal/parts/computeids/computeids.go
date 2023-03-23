package computeids

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/emptybatchid"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/internal/common"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	mppbdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mppbtypes "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	"github.com/filecoin-project/mir/pkg/serializing"
	t "github.com/filecoin-project/mir/pkg/types"
)

// IncludeComputationOfTransactionAndBatchIDs registers event handler for processing RequestTransactionIDs and
// RequestBatchID events.
func IncludeComputationOfTransactionAndBatchIDs(
	m dsl.Module,
	mc *common.ModuleConfig,
	_ *common.ModuleParams,
	_ *common.State,
) {
	mppbdsl.UponRequestTransactionIDs(m, func(txs []*requestpbtypes.Request, origin *mppbtypes.RequestTransactionIDsOrigin) error {
		txMsgs := make([]*commonpbtypes.HashData, len(txs))
		for i, tx := range txs {
			txMsgs[i] = &commonpbtypes.HashData{Data: serializing.RequestForHash(tx.Pb())}
		}

		eventpbdsl.HashRequest(m, mc.Hasher, txMsgs, &computeHashForTransactionIDsContext{origin})
		return nil
	})

	eventpbdsl.UponHashResult(m, func(hashes [][]uint8, context *computeHashForTransactionIDsContext) error {
		txIDs := make([]t.TxID, len(hashes))
		copy(txIDs, hashes)

		mppbdsl.TransactionIDsResponse(m, context.origin.Module, txIDs, context.origin)
		return nil
	})

	mppbdsl.UponRequestBatchID(m, func(txIDs []t.TxID, origin *mppbtypes.RequestBatchIDOrigin) error {
		data := make([][]byte, len(txIDs))
		copy(data, txIDs)

		if len(txIDs) == 0 {
			mppbdsl.BatchIDResponse(m, origin.Module, emptybatchid.EmptyBatchID(), origin)
		}

		dsl.HashOneMessage(m, mc.Hasher, data, &computeHashForBatchIDContext{origin})
		return nil
	})

	dsl.UponOneHashResult(m, func(hash []byte, context *computeHashForBatchIDContext) error {
		mppbdsl.BatchIDResponse(m, context.origin.Module, hash, context.origin)
		return nil
	})
}

// Context data structures

type computeHashForTransactionIDsContext struct {
	origin *mppbtypes.RequestTransactionIDsOrigin
}

type computeHashForBatchIDContext struct {
	origin *mppbtypes.RequestBatchIDOrigin
}
