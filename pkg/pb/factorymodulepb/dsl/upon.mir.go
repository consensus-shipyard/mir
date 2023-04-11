package factorymodulepbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/factorymodulepb/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponFactory[W types.Factory_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Factory](m, func(ev *types.Factory) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponNewModule(m dsl.Module, handler func(moduleId types2.ModuleID, retentionIndex types2.RetentionIndex, params *types.GeneratorParams) error) {
	UponFactory[*types.Factory_NewModule](m, func(ev *types.NewModule) error {
		return handler(ev.ModuleId, ev.RetentionIndex, ev.Params)
	})
}

func UponGarbageCollect(m dsl.Module, handler func(retentionIndex types2.RetentionIndex) error) {
	UponFactory[*types.Factory_GarbageCollect](m, func(ev *types.GarbageCollect) error {
		return handler(ev.RetentionIndex)
	})
}
