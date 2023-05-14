package certcreation

import (
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/common"
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
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// certCreationState represents the state related to this part of the module.
type certCreationState struct {
	nextReqID      requestID
	requestStates  []*requestState
	certificates   map[requestID]*certificate
	remainingSeqNr int
}

// requestID is used to uniquely identify an outgoing request.
type requestID = uint64

// requestState represents the state related to a request on the source node of the request.
// The node disposes of this state as soon as the request is completed.
type requestState struct {
	reqOrigin *apbtypes.RequestCertOrigin
}

type certificate struct {
	batchID     msctypes.BatchID
	sigs        map[t.NodeID][]byte
	receivedSig map[t.NodeID]bool
}

// IncludeCreatingCertificates registers event handlers for processing availabilitypb.RequestCert events.
func IncludeCreatingCertificates(
	m dsl.Module,
	mc common.ModuleConfig,
	params *common.ModuleParams,
	logger logging.Logger,
) {
	state := certCreationState{
		nextReqID:      0,
		requestStates:  make([]*requestState, 0),
		certificates:   make(map[requestID]*certificate),
		remainingSeqNr: params.MaxRequests,
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// code for the source of the request                                                                             //
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	apbdsl.UponComputeCert(m, func() error {
		if state.remainingSeqNr <= 0 {
			return nil //no remaining sequence numbers to process, do not do dumb computations
		}
		reqID := state.nextReqID
		state.nextReqID++
		state.certificates[reqID] = &certificate{
			receivedSig: make(map[t.NodeID]bool),
			sigs:        make(map[t.NodeID][]byte),
		}
		mempooldsl.RequestBatch(m, mc.Mempool, &requestBatchFromMempoolContext{reqID})
		return nil
	})

	// When a batch is requested by the consensus layer, request a batch of transactions from the mempool.
	apbdsl.UponRequestCert(m, func(origin *apbtypes.RequestCertOrigin) error {

		state.requestStates = append(state.requestStates, &requestState{
			reqOrigin: origin,
		})

		respondIfReady(m, &state, params)
		if len(state.certificates) == 0 {
			apbdsl.ComputeCert(m, mc.Self)
		}
		return nil
	})

	// When the mempool provides a batch, compute its ID.
	mempooldsl.UponNewBatch(m, func(txIDs []tt.TxID, txs []*trantorpbtypes.Transaction, context *requestBatchFromMempoolContext) error {
		if len(txs) == 0 {
			// If the batch is empty, immediately stop it and respond to all pending requests with an empty certificate.
			delete(state.certificates, context.reqID)
			respond(m, &state, []*mscpbtypes.Cert{})
			// Note that a creation of another certificate is NOT started here.
			// It will only be triggered by another certificate request (UponRequestCert).
			// This creates a slight inconvenience that, if there is an empty batch no certificates will be pre-computed
			// for the next request, and the next request will have to wait until the availability protocol finishes.
			// TODO: A remedy would be to start a timer here, that would trigger a delayed ComputeCert event
			//   (it's a one-liner here, but the corresponding delay parameter needs to be introduced in the params).
		} else {
			mempooldsl.RequestBatchID(m, mc.Mempool, txIDs, &requestIDOfOwnBatchContext{context.reqID, txs})
		}
		return nil
	})

	// When the id of the batch is computed, request signatures for the batch from all nodes.
	mempooldsl.UponBatchIDResponse(m, func(batchID msctypes.BatchID, context *requestIDOfOwnBatchContext) error {
		if _, ok := state.certificates[context.reqID]; !ok {
			return nil
		}
		state.certificates[context.reqID].batchID = batchID
		// TODO: add persistent storage for crash-recovery.
		transportpbdsl.SendMessage(m, mc.Net, mscpbmsgs.RequestSigMessage(mc.Self, context.txs, context.reqID), params.AllNodes)
		return nil
	})

	// When receive a signature, verify its correctness.
	mscpbdsl.UponSigMessageReceived(m, func(from t.NodeID, signature []byte, reqID requestID) error {
		cert, ok := state.certificates[reqID]
		if !ok {
			// Ignore a message with an invalid or outdated request id.
			return nil
		}

		// check that sender is a member
		if !sliceutil.Contains(params.AllNodes, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}

		if !cert.receivedSig[from] {
			cert.receivedSig[from] = true
			sigData := common.SigData(params.InstanceUID, cert.batchID)
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
		cert, ok := state.certificates[context.reqID]
		if !ok {
			// The request has already been completed.
			return nil
		}

		cert.sigs[nodeID] = context.signature

		newDue := len(cert.sigs) >= params.F+1 // keep this here...

		if len(state.requestStates) > 0 {
			respondIfReady(m, &state, params) // ... because this call changes the state
		}

		if newDue && len(state.certificates) < params.Limit {
			apbdsl.ComputeCert(m, mc.Self)
		}

		return nil
	})

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// code for all other nodes                                                                                       //
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// When receive a request for a signature, compute the ids of the received transactions.
	mscpbdsl.UponRequestSigMessageReceived(m, func(from t.NodeID, txs []*trantorpbtypes.Transaction, reqID requestID) error {
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

func respondIfReady(m dsl.Module, state *certCreationState, params *common.ModuleParams) {

	// Select certificates with enough signatures.
	finishedCerts := maputil.RemoveAll(state.certificates, func(_ requestID, cert *certificate) bool {
		return len(cert.sigs) >= params.F+1
	})

	// Return immediately if there are no finished certificates.
	if len(finishedCerts) == 0 {
		return
	}

	// Convert certificates to the right format for including in a response and send a response to all pending requests.
	responseCerts := sliceutil.Transform(
		maputil.GetValues(finishedCerts),
		func(_ int, cert *certificate) *mscpbtypes.Cert {
			certNodes, certSigs := maputil.GetKeysAndValues(cert.sigs)
			return mscCert(cert.batchID, certNodes, certSigs)
		},
	)
	respond(m, state, responseCerts)
}

func respond(m dsl.Module, state *certCreationState, certs []*mscpbtypes.Cert) {
	// Respond to each pending request
	for _, requestState := range state.requestStates {
		apbdsl.NewCert(m, requestState.reqOrigin.Module, avCert(certs), requestState.reqOrigin)
		state.remainingSeqNr--
	}
	// Dispose of the state associated with handled requests.
	state.requestStates = make([]*requestState, 0)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Protobuf constructors for multisig-collector-specific certificates                                                 //
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// mscCert returns a single multisig (sub)certificate produced by the multisig collector.
// Multiple of these can be aggregated in one overarching certificate used by the availability layer.
func mscCert(batchID msctypes.BatchID, signers []t.NodeID, signatures [][]byte) *mscpbtypes.Cert {
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
	reqID requestID
}

type requestIDOfOwnBatchContext struct {
	reqID requestID
	txs   []*trantorpbtypes.Transaction
}

type computeIDsOfReceivedTxsContext struct {
	sourceID t.NodeID
	txs      []*trantorpbtypes.Transaction
	reqID    requestID
}

type computeIDOfReceivedBatchContext struct {
	sourceID t.NodeID
	txIDs    []tt.TxID
	txs      []*trantorpbtypes.Transaction
	reqID    requestID
}

type storeBatchContext struct {
	sourceID t.NodeID
	reqID    requestID
	batchID  msctypes.BatchID
}

type signReceivedBatchContext struct {
	sourceID t.NodeID
	reqID    requestID
}

type verifySigContext struct {
	reqID     requestID
	signature []byte
}
