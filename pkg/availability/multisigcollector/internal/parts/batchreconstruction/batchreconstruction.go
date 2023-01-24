package batchreconstruction

import (
	"fmt"
	"reflect"

	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/common"
	"github.com/filecoin-project/mir/pkg/dsl"
	mempooldsl "github.com/filecoin-project/mir/pkg/mempool/dsl"
	batchdbpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/dsl"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	mscpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/dsl"
	mscpbmsgs "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/msgs"
	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	apbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// State represents the state related to this part of the module.
type State struct {
	NextReqID    t.RequestID
	RequestState map[t.RequestID]*RequestState
}

// RequestState represents the state related to a request on the source node of the request.
// The node disposes of this state as soon as the request is completed.
type RequestState struct {
	BatchID   t.BatchID
	ReqOrigin *apbtypes.RequestTransactionsOrigin
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
	apbdsl.UponRequestTransactions(m, func(cert *apbtypes.Cert, origin *apbtypes.RequestTransactionsOrigin) error {
		//TODO this switch might be removable with some refactoring
		// NOTE: it is assumed that cert is valid.

		mscCertWrapper, ok := cert.Type.(*apbtypes.Cert_Msc)
		if !ok {
			return fmt.Errorf("unexpected certificate type")
		}
		c := mscCertWrapper.Msc
		batchID := c.BatchId
		batchdbpbdsl.LookupBatch(m, mc.BatchDB, batchID, &lookupBatchLocallyContext{c, origin})
		return nil
	})

	// If the batch is present in the local storage, return it. Otherwise, ask the nodes that signed the certificate.
	batchdbpbdsl.UponLookupBatchResponse(m, func(found bool, txs []*requestpbtypes.Request, metadata []byte, context *lookupBatchLocallyContext) error {
		if found {
			apbdsl.ProvideTransactions(m, context.origin.Module, txs, context.origin)
			return nil
		}

		reqID := state.NextReqID
		state.NextReqID++

		state.RequestState[reqID] = &RequestState{BatchID: context.cert.BatchId, ReqOrigin: context.origin}

		eventpbdsl.SendMessage(m, mc.Net,
			mscpbmsgs.RequestBatchMessage(mc.Self, context.cert.BatchId, reqID),
			context.cert.Signers)
		return nil
	})

	// When receive a request for batch from another node, lookup the batch in the local storage.
	mscpbdsl.UponRequestBatchMessageReceived(m, func(from t.NodeID, batchID t.BatchID, reqID t.RequestID) error {
		// TODO: add some DoS prevention mechanisms.
		batchdbpbdsl.LookupBatch(m, mc.BatchDB, batchID, &lookupBatchOnRemoteRequestContext{from, reqID})
		return nil
	})

	// If the batch is found in the local storage, send it to the requesting node.
	batchdbpbdsl.UponLookupBatchResponse(m, func(found bool, txs []*requestpbtypes.Request, metadata []byte, context *lookupBatchOnRemoteRequestContext) error {
		if !found {
			// Ignore invalid request.
			return nil
		}

		eventpbdsl.SendMessage(m, mc.Net, mscpbmsgs.ProvideBatchMessage(mc.Self, txs, context.reqID), []t.NodeID{context.requester})
		return nil
	})

	// When receive a requested batch, compute the ids of the received transactions.
	mscpbdsl.UponProvideBatchMessageReceived(m, func(from t.NodeID, txs []*requestpbtypes.Request, reqID t.RequestID) error {
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

		if !reflect.DeepEqual(batchID, requestState.BatchID) {
			// The received batch is not valid, keep waiting for a valid response.
			return nil
		}

		batchdbpbdsl.StoreBatch(m, mc.BatchDB, batchID, context.txIDs, context.txs, []byte{} /*metadata*/, &storeBatchContext{})
		apbdsl.ProvideTransactions(m, requestState.ReqOrigin.Module, context.txs, requestState.ReqOrigin)

		// Dispose of the state associated with this request.
		delete(state.RequestState, context.reqID)
		return nil
	})

	batchdbpbdsl.UponBatchStored(m, func(_ *storeBatchContext) error {
		// do nothing.
		return nil
	})
}

// Context data structures

type lookupBatchLocallyContext struct {
	cert   *mscpbtypes.Cert
	origin *apbtypes.RequestTransactionsOrigin
}

type lookupBatchOnRemoteRequestContext struct {
	requester t.NodeID
	reqID     t.RequestID
}

type requestTxIDsContext struct {
	reqID t.RequestID
	txs   []*requestpbtypes.Request
}

type requestBatchIDContext struct {
	reqID t.RequestID
	txs   []*requestpbtypes.Request
	txIDs []t.TxID
}

type storeBatchContext struct{}
