package factorypbevents

import (
	types3 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func NewModule(destModule types.ModuleID, moduleId types.ModuleID, retentionIndex types1.RetentionIndex, params *types2.GeneratorParams) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_Factory{
			Factory: &types2.Event{
				Type: &types2.Event_NewModule{
					NewModule: &types2.NewModule{
						ModuleId:       moduleId,
						RetentionIndex: retentionIndex,
						Params:         params,
					},
				},
			},
		},
	}
}

func GarbageCollect(destModule types.ModuleID, retentionIndex types1.RetentionIndex) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_Factory{
			Factory: &types2.Event{
				Type: &types2.Event_GarbageCollect{
					GarbageCollect: &types2.GarbageCollect{
						RetentionIndex: retentionIndex,
					},
				},
			},
		},
	}
}
