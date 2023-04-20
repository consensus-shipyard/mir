package checkpointpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/checkpointpb/events"
	types2 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func StableCheckpoint(m dsl.Module, destModule types.ModuleID, sn types1.SeqNr, snapshot *types2.StateSnapshot, cert map[types.NodeID][]uint8) {
	dsl.EmitMirEvent(m, events.StableCheckpoint(destModule, sn, snapshot, cert))
}

func EpochProgress(m dsl.Module, destModule types.ModuleID, nodeId types.NodeID, epoch types1.EpochNr) {
	dsl.EmitMirEvent(m, events.EpochProgress(destModule, nodeId, epoch))
}
