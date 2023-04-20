package apppbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/apppb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types4 "github.com/filecoin-project/mir/pkg/trantor/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_App](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponSnapshotRequest(m dsl.Module, handler func(replyTo types2.ModuleID) error) {
	UponEvent[*types.Event_SnapshotRequest](m, func(ev *types.SnapshotRequest) error {
		return handler(ev.ReplyTo)
	})
}

func UponSnapshot(m dsl.Module, handler func(appData []uint8) error) {
	UponEvent[*types.Event_Snapshot](m, func(ev *types.Snapshot) error {
		return handler(ev.AppData)
	})
}

func UponRestoreState(m dsl.Module, handler func(checkpoint *types3.StableCheckpoint) error) {
	UponEvent[*types.Event_RestoreState](m, func(ev *types.RestoreState) error {
		return handler(ev.Checkpoint)
	})
}

func UponNewEpoch(m dsl.Module, handler func(epochNr types4.EpochNr) error) {
	UponEvent[*types.Event_NewEpoch](m, func(ev *types.NewEpoch) error {
		return handler(ev.EpochNr)
	})
}
