package isspbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/isspb/events"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func PushCheckpoint(m dsl.Module, destModule types.ModuleID) {
	dsl.EmitMirEvent(m, events.PushCheckpoint(destModule))
}

func SBDeliver(m dsl.Module, destModule types.ModuleID, sn types.SeqNr, data []uint8, aborted bool, leader types.NodeID, instanceId types.ModuleID) {
	dsl.EmitMirEvent(m, events.SBDeliver(destModule, sn, data, aborted, leader, instanceId))
}
