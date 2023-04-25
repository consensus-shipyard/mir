package batchfetcherpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_BatchFetcher](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponNewOrderedBatch(m dsl.Module, handler func(txs []*types2.Transaction) error) {
	UponEvent[*types.Event_NewOrderedBatch](m, func(ev *types.NewOrderedBatch) error {
		return handler(ev.Txs)
	})
}
