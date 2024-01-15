// Code generated by Mir codegen. DO NOT EDIT.

package synchronizerpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func SyncRequest(m dsl.Module, destModule types.ModuleID, orphanBlock *types1.Block, leaveIds []uint64) {
	dsl.EmitMirEvent(m, events.SyncRequest(destModule, orphanBlock, leaveIds))
}
