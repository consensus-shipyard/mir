package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func Init(m dsl.Module, destModule types.ModuleID) {
	dsl.EmitMirEvent(m, events.Init(destModule))
}

func TimerDelay(m dsl.Module, destModule types.ModuleID, eventsToDelay []*types1.Event, delay types.TimeDuration) {
	dsl.EmitMirEvent(m, events.TimerDelay(destModule, eventsToDelay, delay))
}

func TimerRepeat(m dsl.Module, destModule types.ModuleID, eventsToRepeat []*types1.Event, delay types.TimeDuration, retentionIndex types.RetentionIndex) {
	dsl.EmitMirEvent(m, events.TimerRepeat(destModule, eventsToRepeat, delay, retentionIndex))
}

func TimerGarbageCollect(m dsl.Module, destModule types.ModuleID, retentionIndex types.RetentionIndex) {
	dsl.EmitMirEvent(m, events.TimerGarbageCollect(destModule, retentionIndex))
}
