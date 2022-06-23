package bcb

import (
	"github.com/filecoin-project/mir/pkg/pb/bcbpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

func Message(moduleID t.ModuleID, msg *bcbpb.Message) *messagepb.Message {
	return &messagepb.Message{
		DestModule: moduleID.Pb(),
		Type: &messagepb.Message_Bcb{
			Bcb: msg,
		},
	}
}

func StartMessage(moduleID t.ModuleID, data []byte) *messagepb.Message {
	return Message(moduleID, &bcbpb.Message{
		Type: &bcbpb.Message_StartMessage{
			StartMessage: &bcbpb.StartMessage{Data: data},
		},
	})
}

func EchoMessage(moduleID t.ModuleID, signature []byte) *messagepb.Message {
	return Message(moduleID, &bcbpb.Message{
		Type: &bcbpb.Message_EchoMessage{
			EchoMessage: &bcbpb.EchoMessage{
				Signature: signature,
			},
		},
	})
}

func FinalMessage(moduleID t.ModuleID, data []byte, signers []t.NodeID, signatures [][]byte) *messagepb.Message {
	return Message(moduleID, &bcbpb.Message{
		Type: &bcbpb.Message_FinalMessage{
			FinalMessage: &bcbpb.FinalMessage{
				Data:       data,
				Signers:    t.NodeIDSlicePb(signers),
				Signatures: signatures,
			},
		},
	})
}
