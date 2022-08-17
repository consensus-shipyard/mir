package bcbpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/bcbpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

func BroadcastRequest(next []*types.Event, destModule types1.ModuleID, data []uint8) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Bcb{
			Bcb: &types2.Event{
				Type: &types2.Event_Request{
					Request: &types2.BroadcastRequest{
						Data: data,
					},
				},
			},
		},
	}
}

func Deliver(next []*types.Event, destModule types1.ModuleID, data []uint8) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Bcb{
			Bcb: &types2.Event{
				Type: &types2.Event_Deliver{
					Deliver: &types2.Deliver{
						Data: data,
					},
				},
			},
		},
	}
}
