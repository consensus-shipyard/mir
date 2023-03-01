package checkpointpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/checkpointpb/events"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func StableCheckpoint(m dsl.Module, destModule types.ModuleID, sn types.SeqNr, snapshot *commonpb.StateSnapshot, cert map[string][]uint8) {
	dsl.EmitMirEvent(m, events.StableCheckpoint(destModule, sn, snapshot, cert))
}
