package isspbevents

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func PushCheckpoint(destModule types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Iss{
			Iss: &types2.Event{
				Type: &types2.Event_PushCheckpoint{
					PushCheckpoint: &types2.PushCheckpoint{},
				},
			},
		},
	}
}

func SBDeliver(destModule types.ModuleID, sn types.SeqNr, data []uint8, aborted bool, leader types.NodeID, instanceId types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Iss{
			Iss: &types2.Event{
				Type: &types2.Event_SbDeliver{
					SbDeliver: &types2.SBDeliver{
						Sn:         sn,
						Data:       data,
						Aborted:    aborted,
						Leader:     leader,
						InstanceId: instanceId,
					},
				},
			},
		},
	}
}
