package computeids

import (
	"encoding/binary"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/common"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/emptybatchid"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	mppbdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mppbtypes "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
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
			txMsgs[i] = &commonpbtypes.HashData{Data: serializeRequestForHash(tx.Pb())}
		}

		hasherpbdsl.Request(m, mc.Hasher, txMsgs, &computeHashForTransactionIDsContext{origin})
		return nil
	})

	hasherpbdsl.UponResult(m, func(hashes [][]uint8, context *computeHashForTransactionIDsContext) error {
		txIDs := make([]tt.TxID, len(hashes))
		copy(txIDs, hashes)

		mppbdsl.TransactionIDsResponse(m, context.origin.Module, txIDs, context.origin)
		return nil
	})

	mppbdsl.UponRequestBatchID(m, func(txIDs []tt.TxID, origin *mppbtypes.RequestBatchIDOrigin) error {
		data := make([][]byte, len(txIDs))
		copy(data, txIDs)

		if len(txIDs) == 0 {
			mppbdsl.BatchIDResponse(m, origin.Module, emptybatchid.EmptyBatchID(), origin)
		}

		hasherpbdsl.RequestOne(m, mc.Hasher, &commonpbtypes.HashData{Data: data}, &computeHashForBatchIDContext{origin})
		return nil
	})

	hasherpbdsl.UponResultOne(m, func(hash []byte, context *computeHashForBatchIDContext) error {
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

// Auxiliary functions

func serializeRequestForHash(req *requestpb.Request) [][]byte {
	// Encode integer fields.
	clientIDBuf := []byte(req.ClientId)
	reqNoBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(reqNoBuf, req.ReqNo)

	// Note that the signature is *not* part of the hashed data.

	// Return serialized integers along with the request data itself.
	return [][]byte{clientIDBuf, reqNoBuf, req.Data}
}
