package factorymodulepbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/factorymodulepb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/factorymodulepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func NewModule(m dsl.Module, destModule types.ModuleID, moduleId types.ModuleID, retentionIndex types.RetentionIndex, params *types1.GeneratorParams) {
	dsl.EmitMirEvent(m, events.NewModule(destModule, moduleId, retentionIndex, params))
}

func GarbageCollect(m dsl.Module, destModule types.ModuleID, retentionIndex types.RetentionIndex) {
	dsl.EmitMirEvent(m, events.GarbageCollect(destModule, retentionIndex))
}
