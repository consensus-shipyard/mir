package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
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

func UponSendMessage(m dsl.Module, handler func(msg *types2.Message, destinations []types1.NodeID) error) {
	dsl.UponMirEvent[*types.Event_SendMessage](m, func(ev *types.SendMessage) error {
		return handler(ev.Msg, ev.Destinations)
	})
}

func UponMessageReceived(m dsl.Module, handler func(from types1.NodeID, msg *types2.Message) error) {
	dsl.UponMirEvent[*types.Event_MessageReceived](m, func(ev *types.MessageReceived) error {
		return handler(ev.From, ev.Msg)
	})
}

func UponNewRequests(m dsl.Module, handler func(requests []*types3.Request) error) {
	dsl.UponMirEvent[*types.Event_NewRequests](m, func(ev *types.NewRequests) error {
		return handler(ev.Requests)
	})
}
