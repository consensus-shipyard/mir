// Code generated by Mir codegen. DO NOT EDIT.

package testerpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/testerpb/events"
	stdtypes "github.com/filecoin-project/mir/stdtypes"
)

// Module-specific dsl functions for emitting events.

func Tester(m dsl.Module, destModule stdtypes.ModuleID) {
	dsl.EmitMirEvent(m, events.Tester(destModule))
}
