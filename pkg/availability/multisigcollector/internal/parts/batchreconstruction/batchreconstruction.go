package batchreconstruction

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/emptycert"

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
	decidedOrder []t.BatchIDString
	Txs          map[t.BatchIDString][]*requestpbtypes.Request
	ReqOrigin    *apbtypes.RequestTransactionsOrigin
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

		reqID := state.NextReqID
		state.NextReqID++
		state.RequestState[reqID] = &RequestState{decidedOrder: make([]t.BatchIDString, 0), Txs: make(map[t.BatchIDString][]*requestpbtypes.Request), ReqOrigin: origin}

		mscCertsWrapper, ok := cert.Type.(*apbtypes.Cert_Mscs)
		if !ok {
			return fmt.Errorf("unexpected certificate type")
		}
		cs := mscCertsWrapper.Mscs

		nonEmptyCerts := make([]*mscpbtypes.Cert, 0)
		for _, c := range cs.Certs {
			if !emptycert.IsEmpty(c) {
				nonEmptyCerts = append(nonEmptyCerts, c) // do not add to requests
			}
		}

		if len(nonEmptyCerts) > 0 {
			for _, c := range nonEmptyCerts {
				batchIDString := t.BatchIDString(hex.EncodeToString(c.BatchId))
				state.RequestState[reqID].decidedOrder = append(state.RequestState[reqID].decidedOrder, batchIDString) // save the order in which the batches were decided
				state.RequestState[reqID].Txs[batchIDString] = nil
				batchdbpbdsl.LookupBatch(m, mc.BatchDB, c.BatchId, &lookupBatchLocallyContext{c, origin, reqID})

			}
		} else {
			// give empty batch and delete request
			apbdsl.ProvideTransactions(m, origin.Module, []*requestpbtypes.Request{}, origin)
			delete(state.RequestState, reqID)
		}

		return nil
	})

	// If the batch is present in the local storage, return it. Otherwise, ask the nodes that signed the certificate.
	batchdbpbdsl.UponLookupBatchResponse(m, func(found bool, txs []*requestpbtypes.Request, metadata []byte, context *lookupBatchLocallyContext) error {
		if found {
			if _, ok := state.RequestState[context.reqID]; !ok {
				return nil
			}
			saveAndFinish(m, context.reqID, txs, context.cert.BatchId, context.origin, &state)

		} else {
			eventpbdsl.SendMessage(m, mc.Net,
				mscpbmsgs.RequestBatchMessage(mc.Self, context.cert.BatchId, context.reqID),
				context.cert.Signers)
		}
		return nil
	})

	// When receive a request for batch from another node, lookup the batch in the local storage.
	mscpbdsl.UponRequestBatchMessageReceived(m, func(from t.NodeID, batchID t.BatchID, reqID t.RequestID) error {
		// TODO: add some DoS prevention mechanisms.
		batchdbpbdsl.LookupBatch(m, mc.BatchDB, batchID, &lookupBatchOnRemoteRequestContext{from, reqID, batchID})
		return nil
	})

	// If the batch is found in the local storage, send it to the requesting node.
	batchdbpbdsl.UponLookupBatchResponse(m, func(found bool, txs []*requestpbtypes.Request, metadata []byte, context *lookupBatchOnRemoteRequestContext) error {
		if !found {
			// Ignore invalid request.
			return nil
		}

		eventpbdsl.SendMessage(m, mc.Net, mscpbmsgs.ProvideBatchMessage(mc.Self, txs, context.reqID, context.batchID), []t.NodeID{context.requester})
		return nil
	})

	// When receive a requested batch, compute the ids of the received transactions.
	mscpbdsl.UponProvideBatchMessageReceived(m, func(from t.NodeID, txs []*requestpbtypes.Request, reqID t.RequestID, batchID t.BatchID) error {

		if _, ok := state.RequestState[reqID]; !ok {
			// Ignore a message with an invalid or outdated request id.
			return nil
		}

		mempooldsl.RequestTransactionIDs(m, mc.Mempool, txs, &requestTxIDsContext{reqID, txs, batchID})
		return nil
	})

	// When transaction ids are computed, compute the id of the batch.
	mempooldsl.UponTransactionIDsResponse(m, func(txIDs []t.TxID, context *requestTxIDsContext) error {
		mempooldsl.RequestBatchID(m, mc.Mempool, txIDs, &requestBatchIDContext{context.reqID, context.txs, txIDs, context.batchID})
		return nil
	})

	// When the id of the batch is computed, check if it is correct, store the transactions, and complete the request.
	mempooldsl.UponBatchIDResponse(m, func(batchID t.BatchID, context *requestBatchIDContext) error {
		requestState, ok := state.RequestState[context.reqID]
		if !ok {
			// The request has already been completed.
			return nil
		}

		if !reflect.DeepEqual(batchID, context.batchID) {
			// The received batch is not valid, keep waiting for a valid response.
			return nil
		}

		batchdbpbdsl.StoreBatch(m, mc.BatchDB, batchID, context.txIDs, context.txs, []byte{} /*metadata*/, &storeBatchContext{})
		saveAndFinish(m, context.reqID, context.txs, context.batchID, requestState.ReqOrigin, &state)

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
	reqID  t.RequestID
}

type lookupBatchOnRemoteRequestContext struct {
	requester t.NodeID
	reqID     t.RequestID
	batchID   t.BatchID
}

type requestTxIDsContext struct {
	reqID   t.RequestID
	txs     []*requestpbtypes.Request
	batchID t.BatchID
}

type requestBatchIDContext struct {
	reqID   t.RequestID
	txs     []*requestpbtypes.Request
	txIDs   []t.TxID
	batchID t.BatchID
}

type storeBatchContext struct{}

// saveAndFinish saves the received txs and if all requested batches have been received, it completes the request. Returns a bool indicating if the request has been completed.
func saveAndFinish(m dsl.Module, reqID t.RequestID, txs []*requestpbtypes.Request, batchID t.BatchID, origin *apbtypes.RequestTransactionsOrigin, state *State) bool {
	state.RequestState[reqID].Txs[t.BatchIDString(hex.EncodeToString(batchID))] = txs
	allFound := true
	allTxs := make([]*requestpbtypes.Request, 0)

	for _, batchIDString := range state.RequestState[reqID].decidedOrder {
		txs := state.RequestState[reqID].Txs[batchIDString] // preserve order in which they were decided by the orderer
		if txs == nil {
			allFound = false
			break
		}
		allTxs = append(allTxs, txs...)
		//TODO Optimize here (two different for loops one for checking one for appending)
		//TODO Clean code: have one place instead of duplicating this part of the code in both cases
	}

	if allFound {
		apbdsl.ProvideTransactions(m, origin.Module, allTxs, origin)
		delete(state.RequestState, reqID)
	}
	return allFound
}
