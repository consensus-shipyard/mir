package certverification

import (
	"errors"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/common"
	msctypes "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	"github.com/filecoin-project/mir/pkg/dsl"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	apbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/membutil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// State represents the state related to this part of the module.
type State struct {
	NextReqID    RequestID
	RequestState map[RequestID]*RequestState
}

// RequestID is used to uniquely identify an outgoing request.
type RequestID = uint64

type RequestState struct {
	certVerifiedValid map[msctypes.BatchID]bool
}

// IncludeVerificationOfCertificates registers event handlers for processing availabilitypb.VerifyCert events.
func IncludeVerificationOfCertificates(
	m dsl.Module,
	mc common.ModuleConfig,
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

		// An empty certificate (i.e., one certifying the availability of no transactions) is always valid.
		if len(mscCerts) == 0 {
			apbdsl.CertVerified(m, origin.Module, true, "", origin)
			return nil
		}

		state.RequestState[reqID] = &RequestState{
			certVerifiedValid: make(map[msctypes.BatchID]bool),
		}

		for _, mscCert := range mscCerts {
			state.RequestState[reqID].certVerifiedValid[mscCert.BatchId] = false
			sigData := common.SigData(params.InstanceUID, mscCert.BatchId)
			cryptopbdsl.VerifySigs(m, mc.Crypto,
				/*data*/ sliceutil.Repeat(sigData, len(mscCert.Signers)),
				/*signatures*/ mscCert.Signatures,
				/*nodeIDs*/ mscCert.Signers,
				/*context*/ &verifySigsInCertContext{origin, reqID, mscCert},
			)
		}

		return nil
	})

	// When the signatures in a certificate are verified, output the result of certificate verification.
	cryptopbdsl.UponSigsVerified(m, func(nodeIDs []t.NodeID, errs []error, allOK bool, context *verifySigsInCertContext) error {
		reqID := context.reqID

		if _, ok := state.RequestState[reqID]; !ok {
			return nil
		}

		state.RequestState[reqID].certVerifiedValid[context.cert.BatchId] = allOK
		var err error
		if !allOK {
			err = errors.New("some signatures are invalid")
			valid, errStr := t.ErrorPb(err)
			apbdsl.CertVerified(m, context.origin.Module, valid, errStr, context.origin)
			delete(state.RequestState, reqID)
			return nil
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
		return nil, es.Errorf("the certificate is nil")
	}

	// Check that the certificate is of the right type.
	mscCertsWrapper, ok := cert.Type.(*apbtypes.Cert_Mscs)
	if !ok {
		return nil, es.Errorf("unexpected certificate type")
	}
	mscCerts := mscCertsWrapper.Mscs.Certs

	for _, mscCert := range mscCerts {

		// Check that a quorum of nodes signed the certificate.
		if !membutil.HaveWeakQuorum(params.Membership, mscCert.Signers) {
			return nil, es.Errorf("insufficient weight of signatures: %d, need %d (signers: %v)",
				membutil.WeightOf(params.Membership, mscCert.Signers),
				membutil.WeakQuorum(params.Membership),
				mscCert.Signers,
			)
		}

		if len(mscCert.Signers) != len(mscCert.Signatures) {
			return nil, es.Errorf("the number of signatures does not correspond to the number of signers")
		}

		// Check that the identities of the signing nodes are not repeated.
		alreadySeen := make(map[t.NodeID]struct{})
		for _, nodeID := range mscCert.Signers {
			if _, ok := alreadySeen[nodeID]; ok {
				return nil, es.Errorf("some node ids in the certificate are repeated multiple times")
			}
			alreadySeen[nodeID] = struct{}{}
		}

		// Check that signers are members.
		// TODO: This check can be extremely inefficient (quadratic in membership size), especially at large scale.
		if !sliceutil.ContainsAll(maputil.GetKeys(params.Membership.Nodes), mscCert.Signers) {
			return nil, es.Errorf("certificate contains signatures from non members, signers %v", mscCert.Signers)
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
