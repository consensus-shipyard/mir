package eventpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func Init(destModule types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Init{
			Init: &types1.Init{},
		},
	}
}

func TimerDelay(destModule types.ModuleID, eventsToDelay []*types1.Event, delay types.TimeDuration) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Timer{
			Timer: &types1.TimerEvent{
				Type: &types1.TimerEvent_Delay{
					Delay: &types1.TimerDelay{
						EventsToDelay: eventsToDelay,
						Delay:         delay,
					},
				},
			},
		},
	}
}

func TimerRepeat(destModule types.ModuleID, eventsToRepeat []*types1.Event, delay types.TimeDuration, retentionIndex types.RetentionIndex) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Timer{
			Timer: &types1.TimerEvent{
				Type: &types1.TimerEvent_Repeat{
					Repeat: &types1.TimerRepeat{
						EventsToRepeat: eventsToRepeat,
						Delay:          delay,
						RetentionIndex: retentionIndex,
					},
				},
			},
		},
	}
}

func TimerGarbageCollect(destModule types.ModuleID, retentionIndex types.RetentionIndex) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Timer{
			Timer: &types1.TimerEvent{
				Type: &types1.TimerEvent_GarbageCollect{
					GarbageCollect: &types1.TimerGarbageCollect{
						RetentionIndex: retentionIndex,
					},
				},
			},
		},
	}
}

func AppSnapshotRequest(destModule types.ModuleID, replyTo types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_AppSnapshotRequest{
			AppSnapshotRequest: &types1.AppSnapshotRequest{
				ReplyTo: replyTo,
			},
		},
	}
}

func AppRestoreState(destModule types.ModuleID, checkpoint *types2.StableCheckpoint) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_AppRestoreState{
			AppRestoreState: &types1.AppRestoreState{
				Checkpoint: checkpoint,
			},
		},
	}
}

func NewEpoch(destModule types.ModuleID, epochNr types.EpochNr) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_NewEpoch{
			NewEpoch: &types1.NewEpoch{
				EpochNr: epochNr,
			},
		},
	}
}

func SendMessage(destModule types.ModuleID, msg *types3.Message, destinations []types.NodeID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_SendMessage{
			SendMessage: &types1.SendMessage{
				Msg:          msg,
				Destinations: destinations,
			},
		},
	}
}

func MessageReceived(destModule types.ModuleID, from types.NodeID, msg *types3.Message) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_MessageReceived{
			MessageReceived: &types1.MessageReceived{
				From: from,
				Msg:  msg,
			},
		},
	}
}

func NewRequests(destModule types.ModuleID, requests []*types4.Request) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_NewRequests{
			NewRequests: &types1.NewRequests{
				Requests: requests,
			},
		},
	}
}
