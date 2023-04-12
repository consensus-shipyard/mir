package checkpointpbevents

import (
	types3 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func StableCheckpoint(destModule types.ModuleID, sn types.SeqNr, snapshot *types1.StateSnapshot, cert map[types.NodeID][]uint8) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Checkpoint{
			Checkpoint: &types3.Event{
				Type: &types3.Event_StableCheckpoint{
					StableCheckpoint: &types3.StableCheckpoint{
						Sn:       sn,
						Snapshot: snapshot,
						Cert:     cert,
					},
				},
			},
		},
	}
}

func EpochProgress(destModule types.ModuleID, nodeId types.NodeID, epoch types.EpochNr) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Checkpoint{
			Checkpoint: &types3.Event{
				Type: &types3.Event_EpochProgress{
					EpochProgress: &types3.EpochProgress{
						NodeId: nodeId,
						Epoch:  epoch,
					},
				},
			},
		},
	}
}
