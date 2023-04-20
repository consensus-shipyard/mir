package apppbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/apppb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types4 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func SnapshotRequest(destModule types.ModuleID, replyTo types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_App{
			App: &types2.Event{
				Type: &types2.Event_SnapshotRequest{
					SnapshotRequest: &types2.SnapshotRequest{
						ReplyTo: replyTo,
					},
				},
			},
		},
	}
}

func Snapshot(destModule types.ModuleID, appData []uint8) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_App{
			App: &types2.Event{
				Type: &types2.Event_Snapshot{
					Snapshot: &types2.Snapshot{
						AppData: appData,
					},
				},
			},
		},
	}
}

func RestoreState(destModule types.ModuleID, checkpoint *types3.StableCheckpoint) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_App{
			App: &types2.Event{
				Type: &types2.Event_RestoreState{
					RestoreState: &types2.RestoreState{
						Checkpoint: checkpoint,
					},
				},
			},
		},
	}
}

func NewEpoch(destModule types.ModuleID, epochNr types4.EpochNr) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_App{
			App: &types2.Event{
				Type: &types2.Event_NewEpoch{
					NewEpoch: &types2.NewEpoch{
						EpochNr: epochNr,
					},
				},
			},
		},
	}
}
