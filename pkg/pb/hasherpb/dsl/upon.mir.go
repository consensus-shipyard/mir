package hasherpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Hasher](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponRequest(m dsl.Module, handler func(data []*types.HashData, origin *types.HashOrigin) error) {
	UponEvent[*types.Event_Request](m, func(ev *types.Request) error {
		return handler(ev.Data, ev.Origin)
	})
}

func UponResult[C any](m dsl.Module, handler func(digests [][]uint8, context *C) error) {
	UponEvent[*types.Event_Result](m, func(ev *types.Result) error {
		originWrapper, ok := ev.Origin.Type.(*types.HashOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Digests, context)
	})
}

func UponRequestOne(m dsl.Module, handler func(data *types.HashData, origin *types.HashOrigin) error) {
	UponEvent[*types.Event_RequestOne](m, func(ev *types.RequestOne) error {
		return handler(ev.Data, ev.Origin)
	})
}

func UponResultOne[C any](m dsl.Module, handler func(digest []uint8, context *C) error) {
	UponEvent[*types.Event_ResultOne](m, func(ev *types.ResultOne) error {
		originWrapper, ok := ev.Origin.Type.(*types.HashOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Digest, context)
	})
}
