package batchfetcherpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/events"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func NewOrderedBatch(m dsl.Module, next []*types.Event, destModule types1.ModuleID, txs []*requestpb.Request) {
	dsl.EmitMirEvent(m, events.NewOrderedBatch(next, destModule, txs))
}
