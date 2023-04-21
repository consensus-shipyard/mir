package certcreation

import (
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/common"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/emptycert"
	msctypes "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	batchdbpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/dsl"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	mscpbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/dsl"
	mscpbmsgs "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/msgs"
	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"

	apbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	mempooldsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// State represents the state related to this part of the module.
type State struct {
	NextReqID      RequestID
	RequestStates  []*RequestState
	Certificates   map[RequestID]*Certificate
	RemainingSeqNr int
}

// RequestID is used to uniquely identify an outgoing request.
type RequestID = uint64

// RequestState represents the state related to a request on the source node of the request.
// The node disposes of this state as soon as the request is completed.
type RequestState struct {
	ReqOrigin *apbtypes.RequestCertOrigin
}

type Certificate struct {
	BatchID     msctypes.BatchID
	sigs        map[t.NodeID][]byte
	receivedSig map[t.NodeID]bool
}

// IncludeCreatingCertificates registers event handlers for processing availabilitypb.RequestCert events.
func IncludeCreatingCertificates(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	logger logging.Logger,
) {
	state := State{
		NextReqID:      0,
		RequestStates:  make([]*RequestState, 0),
		Certificates:   make(map[RequestID]*Certificate),
		RemainingSeqNr: params.MaxRequests,
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// code for the source of the request                                                                             //
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	apbdsl.UponComputeCert(m, func() error {
		if state.RemainingSeqNr <= 0 {
			return nil //no remaining sequence numbers to process, do not do dumb computations
		}
		reqID := state.NextReqID
		state.NextReqID++
		state.Certificates[reqID] = &Certificate{
			receivedSig: make(map[t.NodeID]bool),
			sigs:        make(map[t.NodeID][]byte),
		}
		mempooldsl.RequestBatch(m, mc.Mempool, &requestBatchFromMempoolContext{reqID})
		return nil
	})

	// When a batch is requested by the consensus layer, request a batch of transactions from the mempool.
	apbdsl.UponRequestCert(m, func(origin *apbtypes.RequestCertOrigin) error {

		state.RequestStates = append(state.RequestStates, &RequestState{
			ReqOrigin: origin,
		})

		sendIfReady(m, &state, params, false)
		if len(state.Certificates) == 0 {
			apbdsl.ComputeCert(m, mc.Self)
		}
		return nil
	})

	// When the mempool provides a batch, compute its ID.
	mempooldsl.UponNewBatch(m, func(txIDs []tt.TxID, txs []*requestpbtypes.Request, context *requestBatchFromMempoolContext) error {
		if len(txs) == 0 {
			delete(state.Certificates, context.reqID)
			if len(state.RequestStates) > 0 { // do not do work for the empty batch if there is not even a request to answer
				sendIfReady(m, &state, params, true)
				if len(state.Certificates) == 0 {
					apbdsl.ComputeCert(m, mc.Self)
				}
			}
			return nil
		}
		// TODO: add persistent storage for crash-recovery.
		mempooldsl.RequestBatchID(m, mc.Mempool, txIDs, &requestIDOfOwnBatchContext{context.reqID, txs})
		return nil
	})

	// When the id of the batch is computed, request signatures for the batch from all nodes.
	mempooldsl.UponBatchIDResponse(m, func(batchID msctypes.BatchID, context *requestIDOfOwnBatchContext) error {
		if _, ok := state.Certificates[context.reqID]; !ok {
			return nil
		}
		state.Certificates[context.reqID].BatchID = batchID
		// TODO: add persistent storage for crash-recovery.
		transportpbdsl.SendMessage(m, mc.Net, mscpbmsgs.RequestSigMessage(mc.Self, context.txs, context.reqID), params.AllNodes)
		return nil
	})

	// When receive a signature, verify its correctness.
	mscpbdsl.UponSigMessageReceived(m, func(from t.NodeID, signature []byte, reqID RequestID) error {
		certificate, ok := state.Certificates[reqID]
		if !ok {
			// Ignore a message with an invalid or outdated request id.
			return nil
		}

		// check that sender is a member
		if !sliceutil.Contains(params.AllNodes, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}

		if !certificate.receivedSig[from] {
			certificate.receivedSig[from] = true
			sigData := common.SigData(params.InstanceUID, certificate.BatchID)
			cryptopbdsl.VerifySig(m, mc.Crypto, sigData, signature, from, &verifySigContext{reqID, signature})
		}
		return nil
	})

	// When a signature is verified, store it in memory.
	cryptopbdsl.UponSigVerified(m, func(nodeID t.NodeID, err error, context *verifySigContext) error {
		if err != nil {
			// Ignore invalid signature.
			return nil
		}
		certificate, ok := state.Certificates[context.reqID]
		if !ok {
			// The request has already been completed.
			return nil
		}

		certificate.sigs[nodeID] = context.signature

		newDue := len(certificate.sigs) >= params.F+1 // keep this here...

		if len(state.RequestStates) > 0 {
			sendIfReady(m, &state, params, false) // ... because this call changes the state
			// do not send empty certificate because nonempty being computed
		}

		if newDue && len(state.Certificates) < params.Limit {
			apbdsl.ComputeCert(m, mc.Self)
		}

		return nil
	})

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// code for all other nodes                                                                                       //
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// When receive a request for a signature, compute the ids of the received transactions.
	mscpbdsl.UponRequestSigMessageReceived(m, func(from t.NodeID, txs []*requestpbtypes.Request, reqID RequestID) error {
		// check that sender is a member
		if !sliceutil.Contains(params.AllNodes, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		mempooldsl.RequestTransactionIDs(m, mc.Mempool, txs, &computeIDsOfReceivedTxsContext{from, txs, reqID})
		return nil
	})

	// When the ids of the received transactions are computed, compute the id of the batch.
	mempooldsl.UponTransactionIDsResponse(m, func(txIDs []tt.TxID, context *computeIDsOfReceivedTxsContext) error {
		mempooldsl.RequestBatchID(m, mc.Mempool, txIDs,
			&computeIDOfReceivedBatchContext{context.sourceID, txIDs, context.txs, context.reqID})
		return nil
	})

	// When the id of the batch is computed, store the batch persistently.
	mempooldsl.UponBatchIDResponse(m, func(batchID msctypes.BatchID, context *computeIDOfReceivedBatchContext) error {
		batchdbpbdsl.StoreBatch(m, mc.BatchDB, batchID, context.txIDs, context.txs, nil, /*metadata*/
			&storeBatchContext{context.sourceID, context.reqID, batchID})
		// TODO minor optimization: start computing cert without waiting for reply (maybe do not even get a reply from batchdb)
		return nil
	})

	// When the batch is stored, generate a signature
	batchdbpbdsl.UponBatchStored(m, func(context *storeBatchContext) error {
		signedData := common.SigData(params.InstanceUID, context.batchID)
		cryptopbdsl.SignRequest(m, mc.Crypto, signedData, &signReceivedBatchContext{context.sourceID, context.reqID})
		return nil
	})

	// When a signature is generated, send it to the process that sent the request.
	cryptopbdsl.UponSignResult(m, func(signature []byte, context *signReceivedBatchContext) error {
		transportpbdsl.SendMessage(m, mc.Net, mscpbmsgs.SigMessage(mc.Self, signature, context.reqID), []t.NodeID{context.sourceID})
		return nil
	})
}

func sendIfReady(m dsl.Module, state *State, params *common.ModuleParams, sendEmptyIfNoneReady bool) {
	certFinished := make([]*Certificate, 0)
	certsToDelete := make(map[msctypes.BatchIDString]struct{}, 0)
	for _, cert := range state.Certificates {
		if len(cert.sigs) >= params.F+1 { // prepare certificates that are ready
			certFinished = append(certFinished, cert)
			certsToDelete[msctypes.BatchIDString(cert.BatchID)] = struct{}{}
		}
	}

	if len(certFinished) > 0 { // if any certificate is ready
		maputil.FindAndDeleteAll(state.Certificates,
			func(key RequestID, val *Certificate) bool {
				_, ok := certsToDelete[msctypes.BatchIDString(val.BatchID)]
				return ok
			}) // prevent duplicates

		// prepare finished certs for proposal
		proposedCerts := make([]*mscpbtypes.Cert, 0, len(certFinished))
		for _, c := range certFinished { // at least one
			certNodes, certSigs := maputil.GetKeysAndValues(c.sigs)
			cert := mscCert(c.BatchID, certNodes, certSigs)
			proposedCerts = append(proposedCerts, cert)
		}
		proposal := avCert(proposedCerts)

		//return the certificates that are ready to all pending requests
		for _, requestState := range state.RequestStates {
			requestingModule := requestState.ReqOrigin.Module
			apbdsl.NewCert(m, requestingModule, proposal, requestState.ReqOrigin)
			state.RemainingSeqNr--
		}
		// Dispose of the state associated with handled requests.
		state.RequestStates = make([]*RequestState, 0)

	} else if sendEmptyIfNoneReady { // if none are ready and sendEmptyIfNoneReady is true, send empty cert
		emptyProposal := avCert([]*mscpbtypes.Cert{emptycert.EmptyCert()})
		for _, requestState := range state.RequestStates {
			requestingModule := requestState.ReqOrigin.Module
			apbdsl.NewCert(m, requestingModule, emptyProposal, requestState.ReqOrigin)
			state.RemainingSeqNr--
		}
		// Dispose of the state associated with handled requests.
		state.RequestStates = make([]*RequestState, 0)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Protobuf constructors for multisig-collector-specific certificates                                                 //
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// mscCert returns a single multisig (sub)certificate produced by the multisig collector.
// Multiple of these can be aggregated in one overarching certificate used by the availability layer.
func mscCert(batchID []byte, signers []t.NodeID, signatures [][]byte) *mscpbtypes.Cert {
	return &mscpbtypes.Cert{
		BatchId:    batchID,
		Signers:    signers,
		Signatures: signatures,
	}
}

// Creates an availability certificate as output by the availability layer from a list of (sub)certificates
// produced by the multisig collector.
func avCert(certs []*mscpbtypes.Cert) *apbtypes.Cert {
	return &apbtypes.Cert{Type: &apbtypes.Cert_Mscs{
		Mscs: &mscpbtypes.Certs{
			Certs: certs,
		}}}
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
	txIDs    []tt.TxID
	txs      []*requestpbtypes.Request
	reqID    RequestID
}

type storeBatchContext struct {
	sourceID t.NodeID
	reqID    RequestID
	batchID  msctypes.BatchID
}

type signReceivedBatchContext struct {
	sourceID t.NodeID
	reqID    RequestID
}

type verifySigContext struct {
	reqID     RequestID
	signature []byte
}
