package bcbpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/bcbpb/events"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func BroadcastRequest(m dsl.Module, next []*types.Event, destModule types1.ModuleID, data []uint8) {
	dsl.EmitMirEvent(m, events.BroadcastRequest(next, destModule, data))
}

func Deliver(m dsl.Module, next []*types.Event, destModule types1.ModuleID, data []uint8) {
	dsl.EmitMirEvent(m, events.Deliver(next, destModule, data))
}
