package certverification

import (
	"errors"
	"fmt"

	adsl "github.com/filecoin-project/mir/pkg/availability/dsl"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector/internal/common"
	"github.com/filecoin-project/mir/pkg/dsl"
	apb "github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// IncludeVerificationOfCertificates registers event handlers for processing availabilitypb.VerifyCert events.
func IncludeVerificationOfCertificates(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	nodeID t.NodeID,
) {
	// When receive a request to verify a certificate, check that it is structurally correct and verify the signatures.
	adsl.UponVerifyCert(m, func(cert *apb.Cert, origin *apb.VerifyCertOrigin) error {
		mscCert, err := verifyCertificateStructure(params, cert)
		if err != nil {
			adsl.CertVerified(m, t.ModuleID(origin.Module), err, origin)
			return nil
		}

		sigMsg := common.SigData(params.InstanceUID, t.BatchID(mscCert.BatchId))
		dsl.VerifyNodeSigs(m, mc.Crypto,
			/*data*/ sliceutil.Repeat(sigMsg, len(mscCert.Signers)),
			/*signatures*/ mscCert.Signatures,
			/*nodeIDs*/ t.NodeIDSlice(mscCert.Signers),
			/*context*/ &verifySigsInCertContext{origin},
		)
		return nil
	})

	// When the signatures in a certificate are verified, output the result of certificate verification.
	dsl.UponNodeSigsVerified(m, func(nodeIDs []t.NodeID, errs []error, allOK bool, context *verifySigsInCertContext) error {
		var err error
		if !allOK {
			err = errors.New("some signatures are invalid")
		}

		adsl.CertVerified(m, t.ModuleID(context.origin.Module), err, context.origin)
		return nil
	})
}

// verifyCertificateStructure checks that the certificate is well-formed without checking the validity of the
// cryptographic authenticators like hashes and digital signatures.
func verifyCertificateStructure(params *common.ModuleParams, cert *apb.Cert) (*mscpb.Cert, error) {
	// Check that the certificate is present.
	if cert == nil || cert.Type == nil {
		return nil, fmt.Errorf("the certificate is nil")
	}

	// Check that the certificate is of the right type.
	mscCertWrapper, ok := cert.Type.(*apb.Cert_Msc)
	if !ok {
		return nil, fmt.Errorf("unexpected certificate type")
	}
	mscCert := mscCertWrapper.Msc

	// Check that the certificate contains a sufficient number of signatures.
	if len(mscCert.Signers) <= params.F+1 {
		return nil, fmt.Errorf("insuficient number of signatures")
	}

	if len(mscCert.Signers) != len(mscCert.Signatures) {
		return nil, fmt.Errorf("the number of signatures does not correspond to the number of signers")
	}

	// Check that the identities of the signing nodes are not repeated.
	alreadySeen := make(map[t.NodeID]struct{})
	for _, idRaw := range mscCert.Signers {
		id := t.NodeID(idRaw)
		if _, ok := alreadySeen[id]; ok {
			return nil, fmt.Errorf("some node ids in the certificate are repeated multiple times")
		}
		alreadySeen[id] = struct{}{}
	}

	// Check that the identities of the source node and the signing nodes are valid.
	allNodes := make(map[t.NodeID]struct{})
	for _, id := range params.AllNodes {
		allNodes[id] = struct{}{}
	}

	for _, idRaw := range mscCert.Signers {
		if _, ok := allNodes[t.NodeID(idRaw)]; !ok {
			return nil, fmt.Errorf("unknown node id: %v", t.NodeID(idRaw))
		}
	}

	return mscCert, nil
}

// Context data structures                                                                                            //

type verifySigsInCertContext struct {
	origin *apb.VerifyCertOrigin
}
