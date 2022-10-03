package batchfetcherpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/events"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func NewOrderedBatch(m dsl.Module, destModule types.ModuleID, txs []*requestpb.Request) {
	dsl.EmitMirEvent(m, events.NewOrderedBatch(destModule, txs))
}
