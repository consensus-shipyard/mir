package predecisions

import (
	"reflect"

	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/certificates/lightcertificates"
	isspbdsl "github.com/filecoin-project/mir/pkg/pb/isspb/dsl"
	isspbtypes "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"

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

	// Upon predecision received, sign it and broadcast it.
	isspbdsl.UponSBDeliver(m,
		func(sn tt.SeqNr, data []uint8, aborted bool, leader t.NodeID, instanceId t.ModuleID) error {

			if state.Predecided {
				logger.Log(logging.LevelWarn, "Received Predecided event while already Predecided")
				return nil
			}

			state.Predecided = true

			predecision := &isspbtypes.SBDeliver{
				Sn:         sn,
				Data:       data,
				Aborted:    aborted,
				Leader:     leader,
				InstanceId: instanceId,
			}

			serializedPredecision := serializePredecision(predecision)

			state.LocalPredecision = &incommon.LocalPredecision{
				SBDeliver: predecision,
				SignedPredecision: &accpbtypes.SignedPredecision{
					Predecision: serializedPredecision,
				},
			}

			// Sign predecision attaching mc.Self to prevent replay attacks.
			cryptopbdsl.SignRequest(m,
				mc.Crypto,
				&cryptopbtypes.SignedData{
					Data: [][]byte{serializedPredecision}},
				&signRequest{data: serializedPredecision})

			return nil
		})

	cryptopbdsl.UponSignResult(m, func(signature []byte, sr *signRequest) error {
		state.LocalPredecision.SignedPredecision.Signature = signature

		// Broadcast signed predecision to all participants (including oneself).
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

		// Verify signature of received signed predecision.
		cryptopbdsl.VerifySig(m, mc.Crypto,
			&cryptopbtypes.SignedData{Data: [][]byte{predecision, []byte(mc.Self)}},
			signature,
			from,
			&accpbtypes.SignedPredecision{
				Predecision: predecision,
				Signature:   signature})

		return nil
	})

	cryptopbdsl.UponSigVerified(m, func(nodeId t.NodeID, err error, sp *accpbtypes.SignedPredecision) error {
		ApplySigVerified(m, mc, params, state, nodeId, err, sp, true, logger)
		return nil
	})

}

// ApplySigVerified applies the result of a signature verification.
// This function is called once a received signed predecision is verified, but also
// For all signatures of a signed predecision contained in a received certificate
// once they are verified.
func ApplySigVerified(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	state *incommon.State,
	nodeID t.NodeID,
	err error,
	sp *accpbtypes.SignedPredecision,
	flushPoMs bool,
	logger logging.Logger,
) {
	if err != nil {
		logger.Log(logging.LevelDebug, "Signature verification failed")
	}

	// Check if PoM found.
	if state.SignedPredecisions[nodeID] != nil {
		if !reflect.DeepEqual(state.SignedPredecisions[nodeID].Predecision, sp.Predecision) {
			logger.Log(logging.LevelWarn, "Received conflicting signed predecisions from same node")
			// if a PoM for this node has not already been sent.
			if _, ok := state.HandledPoMs[nodeID]; !ok {
				state.UnhandledPoMs = append(state.UnhandledPoMs,
					&accpbtypes.PoM{
						NodeId:           nodeID,
						ConflictingMsg_1: state.SignedPredecisions[nodeID],
						ConflictingMsg_2: sp,
					})

				if flushPoMs {
					poms.HandlePoMs(m, mc, params, state, logger)
				}
			}

			logger.Log(logging.LevelDebug, "Discarding signed predecision as already received one from same node")
		}
	}

	// Store signed predecision.
	state.SignedPredecisions[nodeID] = sp
	state.PredecisionNodeIDs[string(sp.Predecision)] = append(state.PredecisionNodeIDs[string(sp.Predecision)], nodeID)

	// Once verified, if strong quorum, broadcast accpbdsl.FullCertificate.
	if state.DecidedCertificate == nil &&
		membutil.HaveStrongQuorum(params.Membership, state.PredecisionNodeIDs[string(sp.Predecision)]) {
		state.DecidedCertificate = &accpbtypes.FullCertificate{
			Decision: sp.Predecision,
			Signatures: maputil.Transform(
				maputil.Filter(
					state.SignedPredecisions,
					func(
						nodeId t.NodeID,
						predecision *accpbtypes.SignedPredecision,
					) bool {
						return reflect.DeepEqual(predecision.Predecision, sp.Predecision)
					}),
				func(nodeID t.NodeID, sp *accpbtypes.SignedPredecision) (t.NodeID, []byte) {
					return nodeID, sp.Signature
				},
			),
		}

		// Now decision is not just the bytes, need to retrieve actual decision.
		decide(m, mc, params, state, sp.Predecision, logger)

		if params.LightCertificates {
			transportpbdsl.SendMessage(
				m,
				mc.Net,
				accpbmsgs.LightCertificate(mc.Self,
					sp.Predecision),
				maputil.GetKeys(params.Membership.Nodes))
		} else {
			transportpbdsl.SendMessage(
				m,
				mc.Net,
				accpbmsgs.FullCertificate(mc.Self,
					state.DecidedCertificate.Decision,
					state.DecidedCertificate.Signatures),
				maputil.GetKeys(params.Membership.Nodes))
		}
	}

	accpbdsl.UponRequestSBMessageReceived(m, func(from t.NodeID, predecision []byte) error {
		if reflect.DeepEqual(predecision, state.LocalPredecision.SignedPredecision.Predecision) {
			transportpbdsl.SendMessage(m,
				mc.Net,
				accpbmsgs.ProvideSBMessage(mc.Self, state.LocalPredecision.SBDeliver),
				[]t.NodeID{from})
		}
		return nil
	})

	accpbdsl.UponProvideSBMessageReceived(m, func(from t.NodeID, sbDeliver *isspbtypes.SBDeliver) error {
		if state.DecidedCertificate == nil {
			logger.Log(logging.LevelDebug, "Ignoring received SBDeliver message from node %v, no local decision yet", from)
			return nil
		}

		if reflect.DeepEqual(state.DecidedCertificate.Decision, serializePredecision(sbDeliver)) {
			finishWithDecision(m, mc, params, state, sbDeliver, logger)
		}

		return nil
	})
}

func decide(m dsl.Module, mc *common.ModuleConfig, params *common.ModuleParams, state *incommon.State, predecision []byte, logger logging.Logger) {
	// Retrieve actual decision.
	if reflect.DeepEqual(predecision, state.LocalPredecision.SignedPredecision.Predecision) {
		sb := state.LocalPredecision.SBDeliver // convenience variable.
		finishWithDecision(m, mc, params, state, sb, logger)
		return
	}

	// Find the actual predecision from other nodes
	transportpbdsl.SendMessage(
		m,
		mc.Net,
		accpbmsgs.RequestSBMessage(mc.Self,
			predecision),
		state.PredecisionNodeIDs[string(predecision)])

}

func finishWithDecision(
	m dsl.Module,
	mc *common.ModuleConfig,
	params *common.ModuleParams,
	state *incommon.State,
	sb *isspbtypes.SBDeliver,
	logger logging.Logger,
) {

	isspbdsl.SBDeliver(m, mc.App, sb.Sn, sb.Data, sb.Aborted, sb.Leader, sb.InstanceId)
	if params.LightCertificates {
		lightcertificates.ApplyLightCertificatesBuffered(m, mc, state, logger)
	}

}

func serializePredecision(sbDeliver *isspbtypes.SBDeliver) []byte {
	b := make([]byte, 0)
	b = append(b, sbDeliver.Sn.Bytes()...)
	b = append(b, sbDeliver.Data...)

	abortedByte := uint8(0)
	if sbDeliver.Aborted {
		abortedByte = uint8(1)
	}

	b = append(b, abortedByte)
	b = append(b, sbDeliver.Leader...)
	b = append(b, sbDeliver.InstanceId...)
	return b
}

type signRequest struct {
	data []byte
}
