package batchreconstruction

import (
	batchdbdsl "github.com/filecoin-project/mir/pkg/availability/batchdb/dsl"
	adsl "github.com/filecoin-project/mir/pkg/availability/dsl"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/common"
	mscdsl "github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/dsl"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/protobuf"
	"github.com/filecoin-project/mir/pkg/dsl"
	mempooldsl "github.com/filecoin-project/mir/pkg/mempool/dsl"
	apb "github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// State represents the state related to this part of the module.
type State struct {
	NextReqID    RequestID
	RequestState map[RequestID]*RequestState
}

// RequestID is used to uniquely identify an outgoing request.
type RequestID = uint64

// RequestState represents the state related to a request on the source node of the request.
// The node disposes of this state as soon as the request is completed.
type RequestState struct {
	BatchID   t.BatchID
	ReqOrigin *apb.RequestTransactionsOrigin
}

// IncludeBatchReconstruction registers event handlers for processing availabilitypb.RequestTransactions events.
func IncludeBatchReconstruction(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	nodeID t.NodeID,
) {
	state := State{
		NextReqID:    0,
		RequestState: make(map[uint64]*RequestState),
	}

	// When receive a request for transactions, first check the local storage.
	mscdsl.UponRequestTransactions(m, func(cert *mscpb.Cert, origin *apb.RequestTransactionsOrigin) error {
		// NOTE: it is assumed that cert is valid.
		batchdbdsl.LookupBatch(m, mc.BatchDB, t.BatchID(cert.BatchId), &lookupBatchLocallyContext{cert, origin})
		return nil
	})

	// If the batch is present in the local storage, return it. Otherwise, ask the nodes that signed the certificate.
	batchdbdsl.UponLookupBatchResponse(m, func(found bool, txs []*requestpb.Request, metadata []byte, context *lookupBatchLocallyContext) error {
		if found {
			adsl.ProvideTransactions(m, t.ModuleID(context.origin.Module), txs, context.origin)
			return nil
		}

		reqID := state.NextReqID
		state.NextReqID++

		state.RequestState[reqID] = &RequestState{t.BatchID(context.cert.BatchId), context.origin}

		dsl.SendMessage(m, mc.Net,
			protobuf.RequestBatchMessage(mc.Self, t.BatchID(context.cert.BatchId), reqID),
			t.NodeIDSlice(context.cert.Signers))
		return nil
	})

	// When receive a request for batch from another node, lookup the batch in the local storage.
	mscdsl.UponRequestBatchMessageReceived(m, func(from t.NodeID, batchID t.BatchID, reqID RequestID) error {
		// TODO: add some DoS prevention mechanisms.
		batchdbdsl.LookupBatch(m, mc.BatchDB, batchID, &lookupBatchOnRemoteRequestContext{from, reqID})
		return nil
	})

	// If the batch is found in the local storage, send it to the requesting node.
	batchdbdsl.UponLookupBatchResponse(m, func(found bool, txs []*requestpb.Request, metadata []byte, context *lookupBatchOnRemoteRequestContext) error {
		if !found {
			// Ignore invalid request.
			return nil
		}

		dsl.SendMessage(m, mc.Net, protobuf.ProvideBatchMessage(mc.Self, txs, context.reqID), []t.NodeID{context.requester})
		return nil
	})

	// When receive a requested batch, compute the ids of the received transactions.
	mscdsl.UponProvideBatchMessageReceived(m, func(from t.NodeID, txs []*requestpb.Request, reqID RequestID) error {
		_, ok := state.RequestState[reqID]
		if !ok {
			// Ignore a message with an invalid or outdated request id.
			return nil
		}

		mempooldsl.RequestTransactionIDs(m, mc.Mempool, txs, &requestTxIDsContext{reqID, txs})
		return nil
	})

	// When transaction ids are computed, compute the id of the batch.
	mempooldsl.UponTransactionIDsResponse(m, func(txIDs []t.TxID, context *requestTxIDsContext) error {
		mempooldsl.RequestBatchID(m, mc.Mempool, txIDs, &requestBatchIDContext{context.reqID, context.txs, txIDs})
		return nil
	})

	// When the id of the batch is computed, check if it is correct, store the transactions, and complete the request.
	mempooldsl.UponBatchIDResponse(m, func(batchID t.BatchID, context *requestBatchIDContext) error {
		requestState, ok := state.RequestState[context.reqID]
		if !ok {
			// The request has already been completed.
			return nil
		}

		if batchID != requestState.BatchID {
			// The received batch is not valid, keep waiting for a valid response.
			return nil
		}

		batchdbdsl.StoreBatch(m, mc.BatchDB, batchID, context.txIDs, context.txs, []byte{} /*metadata*/, &storeBatchContext{})
		adsl.ProvideTransactions(m, t.ModuleID(requestState.ReqOrigin.Module), context.txs, requestState.ReqOrigin)

		// Dispose of the state associated with this request.
		delete(state.RequestState, context.reqID)
		return nil
	})

	batchdbdsl.UponBatchStored(m, func(_ *storeBatchContext) error {
		// do nothing.
		return nil
	})
}

// Context data structures

type lookupBatchLocallyContext struct {
	cert   *mscpb.Cert
	origin *apb.RequestTransactionsOrigin
}

type lookupBatchOnRemoteRequestContext struct {
	requester t.NodeID
	reqID     RequestID
}

type requestTxIDsContext struct {
	reqID RequestID
	txs   []*requestpb.Request
}

type requestBatchIDContext struct {
	reqID RequestID
	txs   []*requestpb.Request
	txIDs []t.TxID
}

type storeBatchContext struct{}
