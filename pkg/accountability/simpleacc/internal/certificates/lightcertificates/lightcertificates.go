package lightcertificates

import (
	"reflect"

	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/common"
	incommon "github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/common"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	accpbdsl "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/dsl"
	accpbmsgs "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/msgs"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
)

// IncludeLightCertificate implements the (optional) light certificate optimization
// that optimistically sends only the predecision during the light certificate
// so that in the good case where there are no disagreements and all processes
// are correct there is no need to broadcast a full certificate containing O(n) signatures.
func IncludeLightCertificate(m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	state *incommon.State,
	logger logging.Logger,
) {

	lightCertificates := make(map[t.NodeID][]byte)

	accpbdsl.UponLightCertificateReceived(m, func(from t.NodeID, data []byte) error {

		if !params.LightCertificates {
			return nil
		}

		if state.DecidedCertificate == nil {
			logger.Log(logging.LevelDebug, "Received light certificate before decided certificate, buffering it")
			lightCertificates[from] = data
			return nil
		}

		decision := state.DecidedCertificate.Decision

		if !reflect.DeepEqual(decision, data) {
			logger.Log(logging.LevelWarn, "Received light certificate with different predecision than local decision! sending full certificate to node %v", from)
			transportpbdsl.SendMessage(
				m,
				mc.Net,
				accpbmsgs.FullCertificate(mc.Self,
					state.DecidedCertificate.Decision,
					state.DecidedCertificate.Signatures),
				[]t.NodeID{from})
		}
		return nil
	})
}
