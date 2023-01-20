package certcreation

import (
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/common"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/protobuf"
	"github.com/filecoin-project/mir/pkg/dsl"
	mempooldsl "github.com/filecoin-project/mir/pkg/mempool/dsl"
	batchdbpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/dsl"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	mscpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/dsl"
	mscpbmsgs "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/msgs"
	apbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
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
	ReqOrigin *apbtypes.RequestCertOrigin
	BatchID   t.BatchID

	receivedSig map[t.NodeID]bool
	sigs        map[t.NodeID][]byte
}

// IncludeCreatingCertificates registers event handlers for processing availabilitypb.RequestCert events.
func IncludeCreatingCertificates(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	nodeID t.NodeID,
) {
	state := State{
		NextReqID:    0,
		RequestState: make(map[uint64]*RequestState),
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// code for the source of the request                                                                             //
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// When a batch is requested by the consensus layer, request a batch of transactions from the mempool.
	apbdsl.UponRequestCert(m, func(origin *apbtypes.RequestCertOrigin) error {
		reqID := state.NextReqID
		state.NextReqID++

		state.RequestState[reqID] = &RequestState{
			ReqOrigin:   origin,
			receivedSig: make(map[t.NodeID]bool),
			sigs:        make(map[t.NodeID][]byte),
		}

		mempooldsl.RequestBatch(m, mc.Mempool, &requestBatchFromMempoolContext{reqID})
		return nil
	})

	// When the mempool provides a batch, compute its ID.
	mempooldsl.UponNewBatch(m, func(txIDs []t.TxID, txs []*requestpbtypes.Request, context *requestBatchFromMempoolContext) error {
		// TODO: add persistent storage for crash-recovery.
		mempooldsl.RequestBatchID(m, mc.Mempool, txIDs, &requestIDOfOwnBatchContext{context.reqID, txs})
		return nil
	})

	// When the id of the batch is computed, request signatures for the batch from all nodes.
	mempooldsl.UponBatchIDResponse(m, func(batchID t.BatchID, context *requestIDOfOwnBatchContext) error {
		state.RequestState[context.reqID].BatchID = batchID
		// TODO: add persistent storage for crash-recovery.
		eventpbdsl.SendMessage(m, mc.Net, mscpbmsgs.RequestSigMessage(mc.Self, context.txs, context.reqID), params.AllNodes)
		return nil
	})

	// When receive a signature, verify its correctness.
	mscpbdsl.UponSigMessageReceived(m, func(from t.NodeID, signature []byte, reqID RequestID) error {
		requestState, ok := state.RequestState[reqID]
		if !ok {
			// Ignore a message with an invalid or outdated request id.
			return nil
		}

		if !requestState.receivedSig[from] {
			requestState.receivedSig[from] = true
			sigData := common.SigData(params.InstanceUID, requestState.BatchID)
			dsl.VerifyOneNodeSig(m, mc.Crypto, sigData, signature, from, &verifySigContext{reqID, signature})
		}
		return nil
	})

	// When a signature is verified, store it in memory.
	dsl.UponOneNodeSigVerified(m, func(nodeID t.NodeID, err error, context *verifySigContext) error {
		if err != nil {
			// Ignore invalid signature.
			return nil
		}
		requestState, ok := state.RequestState[context.reqID]
		if !ok {
			// The request has already been completed.
			return nil
		}

		requestState.sigs[nodeID] = context.signature
		return nil
	})

	// When F+1 signatures are collected, create and output a certificate.
	dsl.UponCondition(m, func() error {
		// Iterate over active outgoing requests.
		// Most of the time, there is expected to be at most one active outgoing request.
		for reqID, requestState := range state.RequestState {
			if len(requestState.sigs) >= params.F+1 {
				certNodes, certSigs := maputil.GetKeysAndValues(requestState.sigs)

				requestingModule := t.ModuleID(requestState.ReqOrigin.Module)
				cert := apbtypes.CertFromPb(protobuf.Cert(requestState.BatchID, certNodes, certSigs))
				apbdsl.NewCert(m, requestingModule, cert, requestState.ReqOrigin)

				// Dispose of the state associated with this request.
				delete(state.RequestState, reqID)
			}
		}
		return nil
	})

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// code for all other nodes                                                                                       //
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// When receive a request for a signature, compute the ids of the received transactions.
	mscpbdsl.UponRequestSigMessageReceived(m, func(from t.NodeID, txs []*requestpbtypes.Request, reqID RequestID) error {
		mempooldsl.RequestTransactionIDs(m, mc.Mempool, txs, &computeIDsOfReceivedTxsContext{from, txs, reqID})
		return nil
	})

	// When the ids of the received transactions are computed, compute the id of the batch.
	mempooldsl.UponTransactionIDsResponse(m, func(txIDs []t.TxID, context *computeIDsOfReceivedTxsContext) error {
		mempooldsl.RequestBatchID(m, mc.Mempool, txIDs,
			&computeIDOfReceivedBatchContext{context.sourceID, txIDs, context.txs, context.reqID})
		return nil
	})

	// When the id of the batch is computed, store the batch persistently.
	mempooldsl.UponBatchIDResponse(m, func(batchID t.BatchID, context *computeIDOfReceivedBatchContext) error {
		batchdbpbdsl.StoreBatch(m, mc.BatchDB, batchID, context.txIDs, context.txs, nil, /*metadata*/
			&storeBatchContext{context.sourceID, context.reqID, batchID})
		return nil
	})

	// When the batch is stored, generate a signature
	batchdbpbdsl.UponBatchStored(m, func(context *storeBatchContext) error {
		sigMsg := common.SigData(params.InstanceUID, context.batchID)
		dsl.SignRequest(m, mc.Crypto, sigMsg, &signReceivedBatchContext{context.sourceID, context.reqID})
		return nil
	})

	// When a signature is generated, send it to the process that sent the request.
	dsl.UponSignResult(m, func(signature []byte, context *signReceivedBatchContext) error {
		eventpbdsl.SendMessage(m, mc.Net, mscpbmsgs.SigMessage(mc.Self, signature, context.reqID), []t.NodeID{context.sourceID})
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Context data structures                                                                                            //
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type requestBatchFromMempoolContext struct {
	reqID RequestID
}

type requestIDOfOwnBatchContext struct {
	reqID RequestID
	txs   []*requestpbtypes.Request
}

type computeIDsOfReceivedTxsContext struct {
	sourceID t.NodeID
	txs      []*requestpbtypes.Request
	reqID    RequestID
}

type computeIDOfReceivedBatchContext struct {
	sourceID t.NodeID
	txIDs    []t.TxID
	txs      []*requestpbtypes.Request
	reqID    RequestID
}

type storeBatchContext struct {
	sourceID t.NodeID
	reqID    RequestID
	batchID  t.BatchID
}

type signReceivedBatchContext struct {
	sourceID t.NodeID
	reqID    RequestID
}

type verifySigContext struct {
	reqID     RequestID
	signature []byte
}
