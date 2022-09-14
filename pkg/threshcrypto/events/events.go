package events

import (
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	pb "github.com/filecoin-project/mir/pkg/pb/threshcryptopb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// SignShare returns an event representing a request to the threshcrypto module for computing the signature share over data.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// SignShareResult produced by the threshcrypto module.
func SignShare(destModule t.ModuleID, data [][]byte, origin *pb.SignShareOrigin) *eventpb.Event {
	return threshcryptoEvent(destModule, &pb.Event{
		Type: &pb.Event_SignShare{SignShare: &pb.SignShare{
			Data:   data,
			Origin: origin,
		}},
	})
}

// SignShareResult returns an event representing the computation of a signature share by the threshcrypto module.
// It contains the computed signature share and the SignShareOrigin,
// an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained signature.
func SignShareResult(destModule t.ModuleID, signatureShare []byte, origin *pb.SignShareOrigin) *eventpb.Event {
	return threshcryptoEvent(destModule, &pb.Event{Type: &pb.Event_SignShareResult{SignShareResult: &pb.SignShareResult{
		SignatureShare: signatureShare,
		Origin:         origin,
	}}})
}

// VerifyShare returns an event representing a request to the threshcrypto module for verifying
// a signature share over data against the group/module's public key.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// VerifyShareResult produced by the crypto module.
func VerifyShare(destModule t.ModuleID, data [][]byte, sigShare []byte, nodeID t.NodeID, origin *pb.VerifyShareOrigin) *eventpb.Event {
	return threshcryptoEvent(destModule, &pb.Event{
		Type: &pb.Event_VerifyShare{VerifyShare: &pb.VerifyShare{
			Data:           data,
			SignatureShare: sigShare,
			NodeId:         nodeID.Pb(),
			Origin:         origin,
		}},
	})
}

// VerifyShareResult returns an event representing the verification of a signature share by the threshcrypto module.
// It contains the result of the verification (boolean and a string error if applicable) and the VerifyShareOrigin,
// an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained signature.
func VerifyShareResult(destModule t.ModuleID, ok bool, err string, origin *pb.VerifyShareOrigin) *eventpb.Event {
	return threshcryptoEvent(destModule, &pb.Event{Type: &pb.Event_VerifyShareResult{VerifyShareResult: &pb.VerifyShareResult{
		Ok:     ok,
		Error:  err,
		Origin: origin,
	}}})
}

// VerifyFull returns an event representing a request to the threshcrypto module for verifying
// a full signature over data against the group/module's public key.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// VerifyFullResult produced by the crypto module.
func VerifyFull(destModule t.ModuleID, data [][]byte, sigFull []byte, origin *pb.VerifyFullOrigin) *eventpb.Event {
	return threshcryptoEvent(destModule, &pb.Event{
		Type: &pb.Event_VerifyFull{VerifyFull: &pb.VerifyFull{
			Data:          data,
			FullSignature: sigFull,
			Origin:        origin,
		}},
	})
}

// VerifyFullResult returns an event representing the verification of a full signature by the threshcrypto module.
// It contains the result of the verification (boolean and a string error if applicable) and the VerifyFullOrigin,
// an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained signature.
func VerifyFullResult(destModule t.ModuleID, ok bool, err string, origin *pb.VerifyFullOrigin) *eventpb.Event {
	return threshcryptoEvent(destModule, &pb.Event{Type: &pb.Event_VerifyFullResult{VerifyFullResult: &pb.VerifyFullResult{
		Ok:     ok,
		Error:  err,
		Origin: origin,
	}}})
}

// Recover returns an event representing a request to the threshcrypto module for recovering
// a full signature share over data from signature shares.
// The full signature can only be recovered if enough shares are provided, and if they were created from the group's
// private key shares, therefore the full signature is always valid for this data in the group.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// RecoverResult produced by the crypto module.
func Recover(destModule t.ModuleID, data [][]byte, sigShares [][]byte, origin *pb.RecoverOrigin) *eventpb.Event {
	return threshcryptoEvent(destModule, &pb.Event{
		Type: &pb.Event_Recover{Recover: &pb.Recover{
			Data:            data,
			SignatureShares: sigShares,
			Origin:          origin,
		}},
	})
}

// RecoverResult returns an event representing the recovery of a full signature by the threshcrypto module.
// It contains the result of the recovery (boolean, the recovered signature, and a string error if applicable)
// and the RecoverOrigin, an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained signature.
func RecoverResult(destModule t.ModuleID, fullSig []byte, ok bool, err string, origin *pb.RecoverOrigin) *eventpb.Event {
	return threshcryptoEvent(destModule, &pb.Event{Type: &pb.Event_RecoverResult{RecoverResult: &pb.RecoverResult{
		FullSignature: fullSig,
		Ok:            ok,
		Error:         err,
		Origin:        origin,
	}}})
}

func threshcryptoEvent(destModule t.ModuleID, ev *pb.Event) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_ThreshCrypto{
			ThreshCrypto: ev,
		},
	}
}
