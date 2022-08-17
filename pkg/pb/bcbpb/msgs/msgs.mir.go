package bcbpbmsgs

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/bcbpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
)

func StartMessage(destModule string, data []uint8) *types.Message {
	return &types.Message{
		DestModule: destModule,
		Type: &types.Message_Bcb{
			Bcb: &types1.Message{
				Type: &types1.Message_StartMessage{
					StartMessage: &types1.StartMessage{
						Data: data,
					},
				},
			},
		},
	}
}

func EchoMessage(destModule string, signature []uint8) *types.Message {
	return &types.Message{
		DestModule: destModule,
		Type: &types.Message_Bcb{
			Bcb: &types1.Message{
				Type: &types1.Message_EchoMessage{
					EchoMessage: &types1.EchoMessage{
						Signature: signature,
					},
				},
			},
		},
	}
}

func FinalMessage(destModule string, data []uint8, signers []string, signatures [][]uint8) *types.Message {
	return &types.Message{
		DestModule: destModule,
		Type: &types.Message_Bcb{
			Bcb: &types1.Message{
				Type: &types1.Message_FinalMessage{
					FinalMessage: &types1.FinalMessage{
						Data:       data,
						Signers:    signers,
						Signatures: signatures,
					},
				},
			},
		},
	}
}
