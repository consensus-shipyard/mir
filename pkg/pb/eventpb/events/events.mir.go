package eventpbevents

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
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
