package eventpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func SendMessage(destModule types.ModuleID, msg *types1.Message, destinations []types.NodeID) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_SendMessage{
			SendMessage: &types2.SendMessage{
				Msg:          msg,
				Destinations: destinations,
			},
		},
	}
}

func MessageReceived(destModule types.ModuleID, from types.NodeID, msg *types1.Message) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_MessageReceived{
			MessageReceived: &types2.MessageReceived{
				From: from,
				Msg:  msg,
			},
		},
	}
}
