package dsl

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	bfpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

// UponEvent registers a handler for the given availability layer event type.
func UponEvent[EvWrapper bfpb.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponEvent[*eventpb.Event_BatchFetcher](m, func(ev *bfpb.Event) error {
		evWrapper, ok := ev.Type.(EvWrapper)
		if !ok {
			return nil
		}
		return handler(evWrapper.Unwrap())
	})
}

func UponNewOrderedBatch(m dsl.Module, handler func(ev *bfpb.NewOrderedBatch) error) {
	UponEvent[*bfpb.Event_NewOrderedBatch](m, func(ev *bfpb.NewOrderedBatch) error {
		return handler(ev)
	})
}
