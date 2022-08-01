package batchreconstruction

import (
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
	*common.State
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
	commonState *common.State,
) {
	state := State{
		State:        commonState,
		NextReqID:    0,
		RequestState: make(map[uint64]*RequestState),
	}

	// When receive a request for transactions, first check the local storage and then ask other nodes.
	mscdsl.UponRequestTransactions(m, func(cert *mscpb.Cert, origin *apb.RequestTransactionsOrigin) error {
		txIDs, ok := state.BatchStore[t.BatchID(cert.BatchId)]
		if ok {
			txs := make([]*requestpb.Request, len(txIDs))
			for i, txID := range txIDs {
				txs[i] = state.TransactionStore[txID]
			}

			adsl.ProvideTransactions(m, t.ModuleID(origin.Module), txs, origin)
			return nil
		}

		reqID := state.NextReqID
		state.NextReqID++

		state.RequestState[reqID] = &RequestState{t.BatchID(cert.BatchId), origin}

		dsl.SendMessage(m, mc.Net,
			protobuf.RequestBatchMessage(mc.Self, t.BatchID(cert.BatchId), reqID),
			t.NodeIDSlice(cert.Signers))
		return nil
	})

	// When receive a request for batch from another node, send all transactions in response.
	mscdsl.UponRequestBatchMessageReceived(m, func(from t.NodeID, batchID t.BatchID, reqID RequestID) error {
		txIDs, ok := state.BatchStore[batchID]
		if !ok {
			// Ignore invalid request.
			return nil
		}

		txs := make([]*requestpb.Request, len(txIDs))
		for i, txID := range txIDs {
			txs[i] = state.TransactionStore[txID]
		}

		dsl.SendMessage(m, mc.Net, protobuf.ProvideBatchMessage(mc.Self, txs, reqID), []t.NodeID{from})
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

		state.BatchStore[batchID] = context.txIDs
		for i, txID := range context.txIDs {
			state.TransactionStore[txID] = context.txs[i]
		}

		adsl.ProvideTransactions(m, t.ModuleID(requestState.ReqOrigin.Module), context.txs, requestState.ReqOrigin)

		// Dispose of the state associated with this request.
		delete(state.RequestState, context.reqID)
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Context data structures                                                                                            //
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type requestTxIDsContext struct {
	reqID RequestID
	txs   []*requestpb.Request
}

type requestBatchIDContext struct {
	reqID RequestID
	txs   []*requestpb.Request
	txIDs []t.TxID
}
