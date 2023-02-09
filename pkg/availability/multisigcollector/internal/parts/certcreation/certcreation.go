package certcreation

import (
	"encoding/hex"

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
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// State represents the state related to this part of the module.
type State struct {
	NextReqID    RequestID
	RequestState []*RequestState
	Certificate  map[RequestID]*Certificate
}

// RequestID is used to uniquely identify an outgoing request.
type RequestID = uint64

// RequestState represents the state related to a request on the source node of the request.
// The node disposes of this state as soon as the request is completed.
type RequestState struct {
	ReqOrigin *apbtypes.RequestCertOrigin
}

type Certificate struct {
	BatchID     t.BatchID
	sigs        map[t.NodeID][]byte
	receivedSig map[t.NodeID]bool
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
		RequestState: make([]*RequestState, 0),
		Certificate:  make(map[RequestID]*Certificate),
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// code for the source of the request                                                                             //
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	apbdsl.UponComputeCert(m, func() error {
		reqID := state.NextReqID
		state.NextReqID++
		state.Certificate[reqID] = &Certificate{
			receivedSig: make(map[t.NodeID]bool),
			sigs:        make(map[t.NodeID][]byte),
		}
		mempooldsl.RequestBatch(m, mc.Mempool, &requestBatchFromMempoolContext{reqID})
		return nil
	})

	// When a batch is requested by the consensus layer, request a batch of transactions from the mempool.
	apbdsl.UponRequestCert(m, func(origin *apbtypes.RequestCertOrigin) error {

		state.RequestState = append(state.RequestState, &RequestState{
			ReqOrigin: origin,
		})

		sendIfReady(m, &state, params, false)
		if len(state.Certificate) == 0 {
			apbdsl.ComputeCert(m, mc.Self)
			// TODO optimization: stop once as many requests have been answered as there are sequence numbers in a segment
		}
		return nil
	})

	// When the mempool provides a batch, compute its ID.
	mempooldsl.UponNewBatch(m, func(txIDs []t.TxID, txs []*requestpbtypes.Request, context *requestBatchFromMempoolContext) error {
		if len(txs) == 0 {
			delete(state.Certificate, context.reqID)
			if len(state.RequestState) == 0 {
				return nil // do not do work for the empty batch if there is not even a request to answer
			} else {
				sendIfReady(m, &state, params, true)
				if len(state.Certificate) == 0 {
					apbdsl.ComputeCert(m, mc.Self)
					// TODO optimization: stop once as many requests have been answered as there are sequence numbers in a segment
				}
				return nil
			}
		}
		// TODO: add persistent storage for crash-recovery.
		mempooldsl.RequestBatchID(m, mc.Mempool, txIDs, &requestIDOfOwnBatchContext{context.reqID, txs})
		return nil
	})

	// When the id of the batch is computed, request signatures for the batch from all nodes.
	mempooldsl.UponBatchIDResponse(m, func(batchID t.BatchID, context *requestIDOfOwnBatchContext) error {
		if _, ok := state.Certificate[context.reqID]; !ok {
			return nil
		}
		state.Certificate[context.reqID].BatchID = batchID
		// TODO: add persistent storage for crash-recovery.
		eventpbdsl.SendMessage(m, mc.Net, mscpbmsgs.RequestSigMessage(mc.Self, context.txs, context.reqID), params.AllNodes)
		return nil
	})

	// When receive a signature, verify its correctness.
	mscpbdsl.UponSigMessageReceived(m, func(from t.NodeID, signature []byte, reqID RequestID) error {
		certificate, ok := state.Certificate[reqID]
		if !ok {
			// Ignore a message with an invalid or outdated request id.
			return nil
		}

		if !certificate.receivedSig[from] {
			certificate.receivedSig[from] = true
			sigData := common.SigData(params.InstanceUID, certificate.BatchID)
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
		certificate, ok := state.Certificate[context.reqID]
		if !ok {
			// The request has already been completed.
			return nil
		}

		certificate.sigs[nodeID] = context.signature

		newDue := len(certificate.sigs) >= params.F+1 // keep this here...

		if len(state.RequestState) > 0 {
			sendIfReady(m, &state, params, false) // ... because this call changes the state
			// do not send empty certificate because nonempty being computed
		}

		if newDue && uint(len(state.Certificate)) < params.Limit {
			apbdsl.ComputeCert(m, mc.Self)
			// TODO optimization: stop once as many requests have been answered as there are sequence numbers in a segment
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

func sendIfReady(m dsl.Module, state *State, params *common.ModuleParams, sendEmptyIfNoneReady bool) {
	certFinished := make([]*Certificate, 0)
	certsToDelete := make(map[t.BatchIDString]struct{}, 0)
	for _, cert := range state.Certificate {
		if len(cert.sigs) >= params.F+1 { // prepare certificates that are ready
			certFinished = append(certFinished, cert)
			certsToDelete[t.BatchIDString(hex.EncodeToString(cert.BatchID))] = struct{}{}
		}
	}

	if len(certFinished) > 0 { // if any certificate is ready
		maputil.FindAndDeleteAll(state.Certificate,
			func(key RequestID, val *Certificate) bool {
				_, ok := certsToDelete[t.BatchIDString(hex.EncodeToString(val.BatchID))]
				return ok
			}) // prevent duplicates

		// prepare finished certs for proposal
		_proposal := make([]*mscpbtypes.Cert, 0, len(certFinished))
		for _, _cert := range certFinished { // at least one
			certNodes, certSigs := maputil.GetKeysAndValues(_cert.sigs)
			cert := &mscpbtypes.Cert{BatchId: _cert.BatchID, Signers: certNodes, Signatures: certSigs}
			_proposal = append(_proposal, cert)
		}

		proposal := &apbtypes.Cert{Type: &apbtypes.Cert_Mscs{Mscs: &mscpbtypes.Certs{
			Certs: _proposal,
		}}}

		//return the certificates that are ready to all pending requests
		for _, requestState := range state.RequestState {
			requestingModule := requestState.ReqOrigin.Module
			apbdsl.NewCert(m, requestingModule, proposal, requestState.ReqOrigin)
		}
		// Dispose of the state associated with handled requests.
		state.RequestState = make([]*RequestState, 0)

	} else if sendEmptyIfNoneReady { // if none are ready and sendEmptyIfNoneReady is true, send empty cert
		emptyProposal := &apbtypes.Cert{Type: &apbtypes.Cert_Mscs{Mscs: &mscpbtypes.Certs{
			Certs: []*mscpbtypes.Cert{emptycert.EmptyCert()},
		}}}
		for _, requestState := range state.RequestState {
			requestingModule := requestState.ReqOrigin.Module
			apbdsl.NewCert(m, requestingModule, emptyProposal, requestState.ReqOrigin)
		}
		// Dispose of the state associated with handled requests.
		state.RequestState = make([]*RequestState, 0)
	}
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
