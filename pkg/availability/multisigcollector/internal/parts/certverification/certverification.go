package certverification

import (
	"errors"
	"fmt"

	"github.com/filecoin-project/mir/pkg/util/sliceutil"

	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/emptycert"

	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	apbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"

	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/common"
	"github.com/filecoin-project/mir/pkg/dsl"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
)

// State represents the state related to this part of the module.
type State struct {
	NextReqID    RequestID
	RequestState map[RequestID]*RequestState
}

// RequestID is used to uniquely identify an outgoing request.
type RequestID = uint64

type RequestState struct {
	certVerifiedValid map[t.BatchIDString]bool
}

// IncludeVerificationOfCertificates registers event handlers for processing availabilitypb.VerifyCert events.
func IncludeVerificationOfCertificates(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
) {
	state := State{
		NextReqID:    0,
		RequestState: make(map[RequestID]*RequestState),
	}
	// When receive a request to verify a certificate, check that it is structurally correct and verify the signatures.
	apbdsl.UponVerifyCert(m, func(cert *apbtypes.Cert, origin *apbtypes.VerifyCertOrigin) error {

		reqID := state.NextReqID
		state.NextReqID++

		mscCerts, err := verifyCertificateStructure(params, cert)
		valid, errStr := t.ErrorPb(err)

		if err != nil { // invalid certificate, send cert verified event with valid=False
			apbdsl.CertVerified(m, origin.Module, valid, errStr, origin)
			return nil
		}

		state.RequestState[reqID] = &RequestState{
			certVerifiedValid: make(map[t.BatchIDString]bool),
		}

		allEmpty := true
		for _, mscCert := range mscCerts {
			if !emptycert.IsEmpty(mscCert) {
				allEmpty = false
				state.RequestState[reqID].certVerifiedValid[t.BatchIDString(mscCert.BatchId)] = false
				sigMsg := common.SigData(params.InstanceUID, mscCert.BatchId)
				dsl.VerifyNodeSigs(m, mc.Crypto,
					/*data*/ sliceutil.Repeat(sigMsg, len(mscCert.Signers)),
					/*signatures*/ mscCert.Signatures,
					/*nodeIDs*/ mscCert.Signers,
					/*context*/ &verifySigsInCertContext{origin, reqID, mscCert},
				)
			}
		}

		if allEmpty {
			// all certs are empty cert, verify immediately
			apbdsl.CertVerified(m, origin.Module, true, "", origin)
			delete(state.RequestState, reqID)
		}

		return nil
	})

	// When the signatures in a certificate are verified, output the result of certificate verification.
	dsl.UponNodeSigsVerified(m, func(nodeIDs []t.NodeID, errs []error, allOK bool, context *verifySigsInCertContext) error {
		reqID := context.reqID

		if _, ok := state.RequestState[reqID]; !ok {
			return nil
		}

		state.RequestState[reqID].certVerifiedValid[t.BatchIDString(context.cert.BatchId)] = allOK
		var err error
		if !allOK {
			err = errors.New("some signatures are invalid")
			valid, errStr := t.ErrorPb(err)
			apbdsl.CertVerified(m, context.origin.Module, valid, errStr, context.origin)
			delete(state.RequestState, reqID)
		}

		// if we get here, that means so far all received signatures for all certificates in the reqID are verified valid or yet to be verified
		allVerifiedValid := true
		for _, certVerifiedValid := range state.RequestState[reqID].certVerifiedValid {
			if !certVerifiedValid {
				allVerifiedValid = false
				break
			}
		}

		if allVerifiedValid {
			apbdsl.CertVerified(m, context.origin.Module, true, "", context.origin)
			delete(state.RequestState, reqID)
		}
		return nil
	})
}

// verifyCertificateStructure checks that the certificate is well-formed without checking the validity of the
// cryptographic authenticators like hashes and digital signatures.
func verifyCertificateStructure(params *common.ModuleParams, cert *apbtypes.Cert) ([]*mscpbtypes.Cert, error) {
	// Check that the certificate is present.
	if cert == nil || cert.Type == nil {
		return nil, fmt.Errorf("the certificate is nil")
	}

	// Check that the certificate is of the right type.
	mscCertsWrapper, ok := cert.Type.(*apbtypes.Cert_Mscs)
	if !ok {
		return nil, fmt.Errorf("unexpected certificate type")
	}
	mscCerts := mscCertsWrapper.Mscs.Certs

	for _, mscCert := range mscCerts {

		if emptycert.IsEmpty(mscCert) {
			continue
		}
		// Check that the certificate contains a sufficient number of signatures.
		if len(mscCert.Signers) < params.F+1 {
			return nil, fmt.Errorf("insuficient number of signatures: %d, need %d", len(mscCert.Signers), params.F+1)
		}

		if len(mscCert.Signers) != len(mscCert.Signatures) {
			return nil, fmt.Errorf("the number of signatures does not correspond to the number of signers")
		}

		// Check that the identities of the signing nodes are not repeated.
		alreadySeen := make(map[t.NodeID]struct{})
		for _, nodeID := range mscCert.Signers {
			if _, ok := alreadySeen[nodeID]; ok {
				return nil, fmt.Errorf("some node ids in the certificate are repeated multiple times")
			}
			alreadySeen[nodeID] = struct{}{}
		}

		// Check that the identities of the source node and the signing nodes are valid.
		allNodes := make(map[t.NodeID]struct{})
		for _, id := range params.AllNodes {
			allNodes[id] = struct{}{}
		}

		for _, id := range mscCert.Signers {
			if _, ok := allNodes[id]; !ok {
				return nil, fmt.Errorf("unknown node ID: %v", id)
			}
		}
	}
	return mscCerts, nil
}

// Context data structures                                                                                            //

type verifySigsInCertContext struct {
	origin *apbtypes.VerifyCertOrigin
	reqID  RequestID
	cert   *mscpbtypes.Cert
}
