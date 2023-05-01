package apppbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/apppb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func SnapshotRequest(m dsl.Module, destModule types.ModuleID, replyTo types.ModuleID) {
	dsl.EmitMirEvent(m, events.SnapshotRequest(destModule, replyTo))
}

func Snapshot(m dsl.Module, destModule types.ModuleID, appData []uint8) {
	dsl.EmitMirEvent(m, events.Snapshot(destModule, appData))
}

func RestoreState(m dsl.Module, destModule types.ModuleID, checkpoint *types1.StableCheckpoint) {
	dsl.EmitMirEvent(m, events.RestoreState(destModule, checkpoint))
}

func NewEpoch(m dsl.Module, destModule types.ModuleID, epochNr types.EpochNr) {
	dsl.EmitMirEvent(m, events.NewEpoch(destModule, epochNr))
}
