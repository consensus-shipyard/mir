package checkpointpbevents

import (
	types4 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func StableCheckpoint(destModule types.ModuleID, sn types1.SeqNr, snapshot *types2.StateSnapshot, cert map[types.NodeID][]uint8) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_Checkpoint{
			Checkpoint: &types4.Event{
				Type: &types4.Event_StableCheckpoint{
					StableCheckpoint: &types4.StableCheckpoint{
						Sn:       sn,
						Snapshot: snapshot,
						Cert:     cert,
					},
				},
			},
		},
	}
}

func EpochProgress(destModule types.ModuleID, nodeId types.NodeID, epoch types1.EpochNr) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_Checkpoint{
			Checkpoint: &types4.Event{
				Type: &types4.Event_EpochProgress{
					EpochProgress: &types4.EpochProgress{
						NodeId: nodeId,
						Epoch:  epoch,
					},
				},
			},
		},
	}
}
