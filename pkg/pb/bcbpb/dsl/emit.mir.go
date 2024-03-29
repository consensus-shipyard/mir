// Code generated by Mir codegen. DO NOT EDIT.

package bcbpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/bcbpb/events"
	stdtypes "github.com/filecoin-project/mir/stdtypes"
)

// Module-specific dsl functions for emitting events.

func BroadcastRequest(m dsl.Module, destModule stdtypes.ModuleID, data []uint8) {
	dsl.EmitMirEvent(m, events.BroadcastRequest(destModule, data))
}

func Deliver(m dsl.Module, destModule stdtypes.ModuleID, data []uint8) {
	dsl.EmitMirEvent(m, events.Deliver(destModule, data))
}
