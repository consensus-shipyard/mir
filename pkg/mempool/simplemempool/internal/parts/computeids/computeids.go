package computeids

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/common"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	mppbdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mppbtypes "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	"github.com/filecoin-project/mir/pkg/pb/trantorpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/serializing"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// IncludeComputationOfTransactionAndBatchIDs registers event handler for processing RequestTransactionIDs and
// RequestBatchID events.
func IncludeComputationOfTransactionAndBatchIDs(
	m dsl.Module,
	mc common.ModuleConfig,
	_ *common.ModuleParams,
	_ *common.State,
) {
	mppbdsl.UponRequestTransactionIDs(m, func(txs []*trantorpbtypes.Transaction, origin *mppbtypes.RequestTransactionIDsOrigin) error {
		txMsgs := make([]*hasherpbtypes.HashData, len(txs))
		for i, tx := range txs {
			txMsgs[i] = &hasherpbtypes.HashData{Data: serializeTXForHash(tx.Pb())}
		}

		hasherpbdsl.Request(m, mc.Hasher, txMsgs, &computeHashForTransactionIDsContext{origin})
		return nil
	})

	hasherpbdsl.UponResult(m, func(hashes [][]uint8, context *computeHashForTransactionIDsContext) error {
		mppbdsl.TransactionIDsResponse(
			m,
			context.origin.Module,
			sliceutil.Transform(hashes, func(_ int, hash []uint8) tt.TxID {
				return tt.TxID(hash)
			}),
			context.origin,
		)
		return nil
	})

	mppbdsl.UponRequestBatchID(m, func(txIDs []tt.TxID, origin *mppbtypes.RequestBatchIDOrigin) error {
		hasherpbdsl.RequestOne(
			m,
			mc.Hasher,
			&hasherpbtypes.HashData{Data: sliceutil.Transform(txIDs, func(_ int, txId tt.TxID) []byte {
				return []byte(txId)
			})},
			&computeHashForBatchIDContext{origin},
		)
		return nil
	})

	hasherpbdsl.UponResultOne(m, func(hash []byte, context *computeHashForBatchIDContext) error {
		mppbdsl.BatchIDResponse(m, context.origin.Module, tt.TxID(hash), context.origin)
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

// Auxiliary functions

func serializeTXForHash(tx *trantorpb.Transaction) [][]byte {
	// Encode integer fields.
	clientIDBuf := []byte(tx.ClientId)

	// Return serialized integers along with the request data itself.
	return [][]byte{clientIDBuf, serializing.Uint64ToBytes(tx.TxNo), tx.Data}
}
