package batchfetcherpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func NewOrderedBatch(m dsl.Module, destModule types.ModuleID, txs []*types1.Transaction) {
	dsl.EmitMirEvent(m, events.NewOrderedBatch(destModule, txs))
}
