package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponInit(m dsl.Module, handler func() error) {
	dsl.UponMirEvent[*types.Event_Init](m, func(ev *types.Init) error {
		return handler()
	})
}

func UponTimerEvent[W types.TimerEvent_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types.Event_Timer](m, func(ev *types.TimerEvent) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponTimerDelay(m dsl.Module, handler func(eventsToDelay []*types.Event, delay types1.TimeDuration) error) {
	UponTimerEvent[*types.TimerEvent_Delay](m, func(ev *types.TimerDelay) error {
		return handler(ev.EventsToDelay, ev.Delay)
	})
}

func UponTimerRepeat(m dsl.Module, handler func(eventsToRepeat []*types.Event, delay types1.TimeDuration, retentionIndex types1.RetentionIndex) error) {
	UponTimerEvent[*types.TimerEvent_Repeat](m, func(ev *types.TimerRepeat) error {
		return handler(ev.EventsToRepeat, ev.Delay, ev.RetentionIndex)
	})
}

func UponTimerGarbageCollect(m dsl.Module, handler func(retentionIndex types1.RetentionIndex) error) {
	UponTimerEvent[*types.TimerEvent_GarbageCollect](m, func(ev *types.TimerGarbageCollect) error {
		return handler(ev.RetentionIndex)
	})
}

func UponAppSnapshotRequest(m dsl.Module, handler func(replyTo types1.ModuleID) error) {
	dsl.UponMirEvent[*types.Event_AppSnapshotRequest](m, func(ev *types.AppSnapshotRequest) error {
		return handler(ev.ReplyTo)
	})
}

func UponAppRestoreState(m dsl.Module, handler func(checkpoint *types2.StableCheckpoint) error) {
	dsl.UponMirEvent[*types.Event_AppRestoreState](m, func(ev *types.AppRestoreState) error {
		return handler(ev.Checkpoint)
	})
}

func UponNewEpoch(m dsl.Module, handler func(epochNr types1.EpochNr) error) {
	dsl.UponMirEvent[*types.Event_NewEpoch](m, func(ev *types.NewEpoch) error {
		return handler(ev.EpochNr)
	})
}

func UponSendMessage(m dsl.Module, handler func(msg *types3.Message, destinations []types1.NodeID) error) {
	dsl.UponMirEvent[*types.Event_SendMessage](m, func(ev *types.SendMessage) error {
		return handler(ev.Msg, ev.Destinations)
	})
}

func UponMessageReceived(m dsl.Module, handler func(from types1.NodeID, msg *types3.Message) error) {
	dsl.UponMirEvent[*types.Event_MessageReceived](m, func(ev *types.MessageReceived) error {
		return handler(ev.From, ev.Msg)
	})
}

func UponNewRequests(m dsl.Module, handler func(requests []*types4.Request) error) {
	dsl.UponMirEvent[*types.Event_NewRequests](m, func(ev *types.NewRequests) error {
		return handler(ev.Requests)
	})
}
