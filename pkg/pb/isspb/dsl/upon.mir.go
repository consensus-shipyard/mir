package isspbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Iss](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponPushCheckpoint(m dsl.Module, handler func() error) {
	UponEvent[*types.Event_PushCheckpoint](m, func(ev *types.PushCheckpoint) error {
		return handler()
	})
}

func UponSBDeliver(m dsl.Module, handler func(sn types2.SeqNr, data []uint8, aborted bool, leader types2.NodeID, instanceId types2.ModuleID) error) {
	UponEvent[*types.Event_SbDeliver](m, func(ev *types.SBDeliver) error {
		return handler(ev.Sn, ev.Data, ev.Aborted, ev.Leader, ev.InstanceId)
	})
}
