package batchreconstruction

import (
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/common"
	msctypes "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	batchdbpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/dsl"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	mscpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/dsl"
	mscpbmsgs "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/msgs"
	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	apbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	mempooldsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// State represents the state related to this part of the module.
type State struct {
	NextReqID    msctypes.RequestID
	RequestState map[msctypes.RequestID]*RequestState
}

// RequestState represents the state related to a request on the source node of the request.
// The node disposes of this state as soon as the request is completed.
type RequestState struct {
	NonEmptyCerts []msctypes.BatchID
	Txs           map[msctypes.BatchID][]*trantorpbtypes.Transaction
	ReqOrigin     *apbtypes.RequestTransactionsOrigin
}

// IncludeBatchReconstruction registers event handlers for processing availabilitypb.RequestTransactions events.
func IncludeBatchReconstruction(
	m dsl.Module,
	mc common.ModuleConfig,
	params *common.ModuleParams,
	logger logging.Logger,
) {
	state := State{
		NextReqID:    0,
		RequestState: make(map[uint64]*RequestState),
	}

	// When receive a request for transactions, first check the local storage.
	apbdsl.UponRequestTransactions(m, func(cert *apbtypes.Cert, origin *apbtypes.RequestTransactionsOrigin) error {
		//TODO this switch might be removable with some refactoring
		// NOTE: it is assumed that cert is valid.

		mscCertsWrapper, ok := cert.Type.(*apbtypes.Cert_Mscs)
		if !ok {
			return es.Errorf("unexpected certificate type: %T (%v)", cert.Type, cert.Type)
		}
		certs := mscCertsWrapper.Mscs.Certs

		// An empty certificate (i.e., one certifying the availability of no transactions)
		// corresponds to an empty list of transactions. Respond immediately.
		if len(certs) == 0 {
			apbdsl.ProvideTransactions(m, origin.Module, []*trantorpbtypes.Transaction{}, origin)
			return nil
		}

		reqID := state.NextReqID
		state.NextReqID++
		state.RequestState[reqID] = &RequestState{
			NonEmptyCerts: make([]msctypes.BatchID, 0),
			Txs:           make(map[msctypes.BatchID][]*trantorpbtypes.Transaction),
			ReqOrigin:     origin,
		}

		for _, c := range certs {
			state.RequestState[reqID].NonEmptyCerts = append(state.RequestState[reqID].NonEmptyCerts, c.BatchId) // save the order in which the batches were decided
			state.RequestState[reqID].Txs[c.BatchId] = nil
			batchdbpbdsl.LookupBatch(m, mc.BatchDB, c.BatchId, &lookupBatchLocallyContext{c, origin, reqID})
		}

		return nil
	})

	// If the batch is present in the local storage, return it. Otherwise, ask the nodes that signed the certificate.
	batchdbpbdsl.UponLookupBatchResponse(m, func(found bool, txs []*trantorpbtypes.Transaction, context *lookupBatchLocallyContext) error {

		if _, ok := state.RequestState[context.reqID]; !ok {
			return nil
		}

		if found {
			saveAndFinish(m, context.reqID, txs, context.cert.BatchId, context.origin, &state)

		} else {
			// TODO optimize: do not ask everyone at once
			// TODO periodically send requests if nodes do not respond
			transportpbdsl.SendMessage(m, mc.Net,
				mscpbmsgs.RequestBatchMessage(mc.Self, context.cert.BatchId, context.reqID),
				context.cert.Signers)
		}
		return nil
	})

	// When receive a request for batch from another node, lookup the batch in the local storage.
	mscpbdsl.UponRequestBatchMessageReceived(m, func(from t.NodeID, batchID msctypes.BatchID, reqID msctypes.RequestID) error {
		// check that sender is a member
		if _, ok := params.Membership.Nodes[from]; !ok {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		// TODO: add some DoS prevention mechanisms.
		batchdbpbdsl.LookupBatch(m, mc.BatchDB, batchID, &lookupBatchOnRemoteRequestContext{from, reqID, batchID})
		return nil
	})

	// If the batch is found in the local storage, send it to the requesting node.
	batchdbpbdsl.UponLookupBatchResponse(m, func(found bool, txs []*trantorpbtypes.Transaction, context *lookupBatchOnRemoteRequestContext) error {
		if !found {
			// Ignore invalid request.
			return nil
		}

		transportpbdsl.SendMessage(m, mc.Net, mscpbmsgs.ProvideBatchMessage(mc.Self, txs, context.reqID, context.batchID), []t.NodeID{context.requester})
		return nil
	})

	// When receive a requested batch, compute the ids of the received transactions.
	mscpbdsl.UponProvideBatchMessageReceived(m, func(from t.NodeID, txs []*trantorpbtypes.Transaction, reqID msctypes.RequestID, batchID msctypes.BatchID) error {
		// check that sender is a member
		if _, ok := params.Membership.Nodes[from]; !ok {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}

		if req, ok := state.RequestState[reqID]; !ok {
			// Ignore a message with an invalid or outdated request id.
			return nil
		} else if _, ok = req.Txs[batchID]; !ok {
			logger.Log(logging.LevelWarn, "Ignoring received batch %v that was not requested.\n", batchID)
			return nil
		}

		if len(txs) == 0 {
			logger.Log(logging.LevelWarn, "Ignoring empty batch %v.\n")
			return nil
		}

		mempooldsl.RequestTransactionIDs(m, mc.Mempool, txs, &requestTxIDsContext{reqID, txs, batchID})
		return nil
	})

	// When transaction ids are computed, compute the id of the batch.
	mempooldsl.UponTransactionIDsResponse(m, func(txIDs []tt.TxID, context *requestTxIDsContext) error {
		mempooldsl.RequestBatchID(m, mc.Mempool, txIDs, &requestBatchIDContext{context.reqID, context.txs, txIDs, context.batchID})
		return nil
	})

	// When the id of the batch is computed, check if it is correct, store the transactions, and complete the request.
	mempooldsl.UponBatchIDResponse(m, func(batchID msctypes.BatchID, context *requestBatchIDContext) error {
		requestState, ok := state.RequestState[context.reqID]
		if !ok {
			// The request has already been completed.
			return nil
		}

		if batchID != context.batchID {
			// The received batch is not valid, keep waiting for a valid response.
			return nil
		}

		batchdbpbdsl.StoreBatch(m,
			mc.BatchDB,
			batchID,
			context.txs,
			tt.RetentionIndex(params.EpochNr),
			&storeBatchContext{},
		)
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
	reqID  msctypes.RequestID
}

type lookupBatchOnRemoteRequestContext struct {
	requester t.NodeID
	reqID     msctypes.RequestID
	batchID   msctypes.BatchID
}

type requestTxIDsContext struct {
	reqID   msctypes.RequestID
	txs     []*trantorpbtypes.Transaction
	batchID msctypes.BatchID
}

type requestBatchIDContext struct {
	reqID   msctypes.RequestID
	txs     []*trantorpbtypes.Transaction
	txIDs   []tt.TxID
	batchID msctypes.BatchID
}

type storeBatchContext struct{}

// saveAndFinish saves the received txs and if all requested batches have been received, it completes the request. Returns a bool indicating if the request has been completed.
func saveAndFinish(m dsl.Module, reqID msctypes.RequestID, txs []*trantorpbtypes.Transaction, batchID msctypes.BatchID, origin *apbtypes.RequestTransactionsOrigin, state *State) bool {
	state.RequestState[reqID].Txs[batchID] = txs
	allFound := true
	allTxs := make([]*trantorpbtypes.Transaction, 0)

	for _, bID := range state.RequestState[reqID].NonEmptyCerts {
		txsSorted := state.RequestState[reqID].Txs[bID] // preserve order in which they were decided by the orderer
		if txsSorted == nil {
			allFound = false
			break
		}
		allTxs = append(allTxs, txsSorted...)
	}

	if allFound {
		apbdsl.ProvideTransactions(m, origin.Module, allTxs, origin)
		delete(state.RequestState, reqID)
	}
	return allFound
}
