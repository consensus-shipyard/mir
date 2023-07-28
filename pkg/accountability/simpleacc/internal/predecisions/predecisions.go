package predecisions

import (
	"reflect"

	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/common"
	incommon "github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/common"
	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/poms"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	accpbdsl "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/dsl"
	accpbmsgs "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/msgs"
	accpbtypes "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/types"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/membutil"
)

// IncludePredecisions implements the broadcast and treatment of predecisions.
// The broadcast and treatment of certificates is
// externalized to the ../certificates package so that a system can decide whether to instantiate the accountability
// module with full certificates or with the optimistic light certificates.
// The broadcast and treatment of PoMs is also externalized to the ../poms package.
func IncludePredecisions(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	state *incommon.State,
	logger logging.Logger,
) {

	accpbdsl.UponPredecided(m, func(data []byte) error {

		if state.Predecided {
			logger.Log(logging.LevelWarn, "Received Predecided message while already Predecided")
			return nil
		}

		state.Predecided = true

		// Sign predecision attaching mc.Self to prevent replay attacks
		cryptopbdsl.SignRequest(m, mc.Crypto, &cryptopbtypes.SignedData{[][]byte{data, []byte(mc.Self)}}, &signRequest{data})

		return nil
	})

	cryptopbdsl.UponSignResult(m, func(signature []byte, sr *signRequest) error {
		state.SignedPredecision = &accpbtypes.SignedPredecision{
			Predecision: sr.data,
			Signature:   signature,
		}

		// Broadcast signed predecision to all participants (including oneself)
		transportpbdsl.SendMessage(m, mc.Net, accpbmsgs.SignedPredecision(mc.Self, sr.data, signature), maputil.GetKeys(params.Membership.Nodes))
		return nil
	})

	accpbdsl.UponSignedPredecisionReceived(m, func(from t.NodeID, predecision []byte, signature []byte) error {

		if state.SignedPredecisions[from] != nil {
			if reflect.DeepEqual(state.SignedPredecisions[from].Predecision, predecision) {
				logger.Log(logging.LevelDebug, "Received the same predecision from node %v twice, ignoring", from)
				return nil
			}
		}

		// Verify signature of received signed predecision
		cryptopbdsl.VerifySig(m, mc.Crypto,
			&cryptopbtypes.SignedData{[][]byte{predecision, []byte(mc.Self)}},
			signature,
			from,
			&accpbtypes.SignedPredecision{
				Predecision: predecision,
				Signature:   signature})

		return nil
	})

	cryptopbdsl.UponSigVerified(m, func(nodeId t.NodeID, err error, sp *accpbtypes.SignedPredecision) error {
		return ApplySigVerified(m, mc, params, state, nodeId, err, sp, true, logger)
	})

}

// ApplySigVerified applies the result of a signature verification
// This function is called once a received signed predecision is verified, but also
// For all signatures of a signed predecision contained in a received certificate
// once they are verified
func ApplySigVerified(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	state *incommon.State,
	nodeId t.NodeID,
	err error,
	sp *accpbtypes.SignedPredecision,
	flushPoMs bool,
	logger logging.Logger,
) error {
	if err != nil {
		logger.Log(logging.LevelDebug, "Signature verification failed")
		return nil
	}

	// Check if PoM found
	if state.SignedPredecisions[nodeId] != nil {
		if !reflect.DeepEqual(state.SignedPredecisions[nodeId].Predecision, sp.Predecision) {
			logger.Log(logging.LevelWarn, "Received conflicting signed predecisions from same node")
			// if a PoM for this node has not already been sent
			if _, ok := state.SentPoMs[nodeId]; !ok {
				state.UnsentPoMs = append(state.UnsentPoMs,
					&accpbtypes.PoM{
						NodeId:           nodeId,
						ConflictingMsg_1: state.SignedPredecisions[nodeId],
						ConflictingMsg_2: sp,
					})

				if flushPoMs {
					poms.SendPoMs(m, mc, params, state, logger)
				}
			}

			logger.Log(logging.LevelDebug, "Discarding signed predecision as already received one from same node")
			return nil
		}
	}

	// Store signed predecision
	state.SignedPredecisions[nodeId] = sp
	if state.PredecisionCount[string(sp.Predecision)] == nil {
		state.PredecisionCount[string(sp.Predecision)] = make([]t.NodeID, 0)
	}
	state.PredecisionCount[string(sp.Predecision)] = append(state.PredecisionCount[string(sp.Predecision)], nodeId)

	// Once verified, if strong quorum, broadcast accpbdsl.FullCertificate
	if state.DecidedCertificate == nil &&
		membutil.HaveStrongQuorum(params.Membership, state.PredecisionCount[string(sp.Predecision)]) {
		state.DecidedCertificate = maputil.Filter(
			state.SignedPredecisions,
			func(
				nodeId t.NodeID,
				predecision *accpbtypes.SignedPredecision,
			) bool {
				return reflect.DeepEqual(predecision.Predecision, sp.Predecision)
			})

		accpbdsl.Decided(m, mc.App, sp.Predecision)

		transportpbdsl.SendMessage(
			m,
			mc.Net,
			accpbmsgs.FullCertificate(mc.Self,
				state.DecidedCertificate),
			maputil.GetKeys(params.Membership.Nodes))
	}

	return nil
}

type signRequest struct {
	data []byte
}
