// Code generated by Mir codegen. DO NOT EDIT.

package bcbpbmsgs

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/bcbpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	stdtypes "github.com/filecoin-project/mir/stdtypes"
)

func StartMessage(destModule stdtypes.ModuleID, data []uint8) *types.Message {
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

func EchoMessage(destModule stdtypes.ModuleID, signature []uint8) *types.Message {
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

func FinalMessage(destModule stdtypes.ModuleID, data []uint8, signers []stdtypes.NodeID, signatures [][]uint8) *types.Message {
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
