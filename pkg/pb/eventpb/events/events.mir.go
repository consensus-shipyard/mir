package eventpbevents

import (
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

func MessageReceived(next []*types.Event, destModule types1.ModuleID, from types1.NodeID, msg *types2.Message) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_MessageReceived{
			MessageReceived: &types.MessageReceived{
				From: from,
				Msg:  msg,
			},
		},
	}
}
