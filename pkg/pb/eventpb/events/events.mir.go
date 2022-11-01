package eventpbevents

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func SignRequest(destModule types.ModuleID, data [][]uint8, origin *types1.SignOrigin) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_SignRequest{
			SignRequest: &types1.SignRequest{
				Data:   data,
				Origin: origin,
			},
		},
	}
}

func SignResult(destModule types.ModuleID, signature []uint8, origin *types1.SignOrigin) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_SignResult{
			SignResult: &types1.SignResult{
				Signature: signature,
				Origin:    origin,
			},
		},
	}
}

func VerifyNodeSigs(destModule types.ModuleID, data []*types1.SigVerData, signatures [][]uint8, origin *types1.SigVerOrigin, nodeIds []types.NodeID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_VerifyNodeSigs{
			VerifyNodeSigs: &types1.VerifyNodeSigs{
				Data:       data,
				Signatures: signatures,
				Origin:     origin,
				NodeIds:    nodeIds,
			},
		},
	}
}

func NodeSigsVerified(destModule types.ModuleID, origin *types1.SigVerOrigin, nodeIds []types.NodeID, valid []bool, errors []error, allOk bool) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_NodeSigsVerified{
			NodeSigsVerified: &types1.NodeSigsVerified{
				Origin:  origin,
				NodeIds: nodeIds,
				Valid:   valid,
				Errors:  errors,
				AllOk:   allOk,
			},
		},
	}
}

func SendMessage(destModule types.ModuleID, msg *types2.Message, destinations []types.NodeID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_SendMessage{
			SendMessage: &types1.SendMessage{
				Msg:          msg,
				Destinations: destinations,
			},
		},
	}
}

func MessageReceived(destModule types.ModuleID, from types.NodeID, msg *types2.Message) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_MessageReceived{
			MessageReceived: &types1.MessageReceived{
				From: from,
				Msg:  msg,
			},
		},
	}
}
