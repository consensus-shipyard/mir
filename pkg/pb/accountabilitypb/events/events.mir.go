// Code generated by Mir codegen. DO NOT EDIT.

package accountabilitypbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func Predecided(destModule types.ModuleID, data []uint8) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Accountability{
			Accountability: &types2.Event{
				Type: &types2.Event_Predecided{
					Predecided: &types2.Predecided{
						Data: data,
					},
				},
			},
		},
	}
}

func Decided(destModule types.ModuleID, data []uint8) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Accountability{
			Accountability: &types2.Event{
				Type: &types2.Event_Decided{
					Decided: &types2.Decided{
						Data: data,
					},
				},
			},
		},
	}
}

func PoMs(destModule types.ModuleID, poms []*types2.PoM) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Accountability{
			Accountability: &types2.Event{
				Type: &types2.Event_Poms{
					Poms: &types2.PoMs{
						Poms: poms,
					},
				},
			},
		},
	}
}
