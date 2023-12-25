package fullcertificates

import (
	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/common"
	incommon "github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/common"
	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/poms"
	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/predecisions"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	accpbdsl "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/dsl"
	accpbtypes "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/types"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/membutil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// IncludeFullCertificate implements the full certificate brodcast and verification
// in order to find PoMs.
func IncludeFullCertificate(m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	state *incommon.State,
	logger logging.Logger,
) {

	accpbdsl.UponFullCertificateReceived(m, func(from t.NodeID, decision []byte, certificate map[t.NodeID][]byte) error {
		if len(certificate) == 0 {
			logger.Log(logging.LevelDebug, "Ignoring empty predecision certificate")
			return nil
		}

		if !membutil.HaveStrongQuorum(params.Membership, maputil.GetKeys(certificate)) {
			logger.Log(logging.LevelDebug, "Ignoring predecision certificate without strong quorum")
			return nil
		}

		// Verify all signatures in certificate.
		cryptopbdsl.VerifySigs(
			m,
			mc.Crypto,
			sliceutil.Generate(
				len(certificate),
				func(i int) *cryptopbtypes.SignedData {
					return &cryptopbtypes.SignedData{
						Data: [][]byte{decision},
					}
				}),
			maputil.GetValues(certificate),
			maputil.GetKeys(certificate),
			&verifySigs{
				certificate: &accpbtypes.FullCertificate{
					Decision:   decision,
					Signatures: certificate,
				},
			},
		)
		return nil
	})

	cryptopbdsl.UponSigsVerified(m, func(nodeIds []t.NodeID, errs []error, allOk bool, vsr *verifySigs) error {
		for i, nodeID := range nodeIds {
			sp := &accpbtypes.SignedPredecision{
				Predecision: vsr.certificate.Decision,
				Signature:   vsr.certificate.Signatures[nodeID],
			}
			predecisions.ApplySigVerified(m, mc, params, state, nodeID, errs[i], sp, false, logger)
		}
		poms.HandlePoMs(m, mc, params, state, logger)
		return nil
	})
}

type verifySigs struct {
	certificate *accpbtypes.FullCertificate
}
