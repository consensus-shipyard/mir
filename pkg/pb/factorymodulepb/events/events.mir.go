package factorymodulepbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/factorymodulepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func NewModule(destModule types.ModuleID, moduleId types.ModuleID, retentionIndex types.RetentionIndex, params *types1.GeneratorParams) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Factory{
			Factory: &types1.Factory{
				Type: &types1.Factory_NewModule{
					NewModule: &types1.NewModule{
						ModuleId:       moduleId,
						RetentionIndex: retentionIndex,
						Params:         params,
					},
				},
			},
		},
	}
}

func GarbageCollect(destModule types.ModuleID, retentionIndex types.RetentionIndex) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Factory{
			Factory: &types1.Factory{
				Type: &types1.Factory_GarbageCollect{
					GarbageCollect: &types1.GarbageCollect{
						RetentionIndex: retentionIndex,
					},
				},
			},
		},
	}
}
