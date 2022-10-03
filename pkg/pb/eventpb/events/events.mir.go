package eventpbevents

import (
	eventpb "github.com/filecoin-project/mir/pkg/pb/eventpb"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func NodeSigsVerified(destModule types.ModuleID, origin *eventpb.SigVerOrigin, nodeIds []types.NodeID, valid []bool, errors []error, allOk bool) *types1.Event {
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
