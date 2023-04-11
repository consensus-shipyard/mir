package checkpointpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/checkpointpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func StableCheckpoint(m dsl.Module, destModule types.ModuleID, sn types.SeqNr, snapshot *types1.StateSnapshot, cert map[types.NodeID][]uint8) {
	dsl.EmitMirEvent(m, events.StableCheckpoint(destModule, sn, snapshot, cert))
}

func EpochProgress(m dsl.Module, destModule types.ModuleID, nodeId types.NodeID, epoch types.EpochNr) {
	dsl.EmitMirEvent(m, events.EpochProgress(destModule, nodeId, epoch))
}
