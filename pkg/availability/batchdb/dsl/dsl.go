package dsl

import (
	batchdbevents "github.com/filecoin-project/mir/pkg/availability/batchdb/events"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

// LookupBatch is used to pull a batch with its metadata from the local batch database.
func LookupBatch[C any](m dsl.Module, dest t.ModuleID, batchID t.BatchID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &batchdbpb.LookupBatchOrigin{
		Module: m.ModuleID().Pb(),
		Type: &batchdbpb.LookupBatchOrigin_Dsl{
			Dsl: dsl.Origin(contextID),
		},
	}

	dsl.EmitEvent(m, batchdbevents.LookupBatch(dest, batchID, origin))
}

// LookupBatchResponse is a response to a LookupBatch event.
func LookupBatchResponse(m dsl.Module, dest t.ModuleID, found bool, txs []*requestpb.Request, metadata []byte, origin *batchdbpb.LookupBatchOrigin) {
	dsl.EmitEvent(m, batchdbevents.LookupBatchResponse(dest, found, txs, metadata, origin))
}

// StoreBatch is used to store a new batch in the local batch database.
func StoreBatch[C any](m dsl.Module, dest t.ModuleID, batchID t.BatchID, txIDs []t.TxID, txs []*requestpb.Request, metadata []byte, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &batchdbpb.StoreBatchOrigin{
		Module: m.ModuleID().Pb(),
		Type: &batchdbpb.StoreBatchOrigin_Dsl{
			Dsl: dsl.Origin(contextID),
		},
	}

	dsl.EmitEvent(m, batchdbevents.StoreBatch(dest, batchID, txIDs, txs, metadata, origin))
}

// BatchStored is a response to a StoreBatch event.
func BatchStored(m dsl.Module, dest t.ModuleID, origin *batchdbpb.StoreBatchOrigin) {
	dsl.EmitEvent(m, batchdbevents.BatchStored(dest, origin))
}

// Module-specific dsl functions for processing events.

// UponEvent registers a handler for the given batchdb event type.
func UponEvent[EvWrapper batchdbpb.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponEvent[*eventpb.Event_BatchDb](m, func(ev *batchdbpb.Event) error {
		evWrapper, ok := ev.Type.(EvWrapper)
		if !ok {
			return nil
		}
		return handler(evWrapper.Unwrap())
	})
}

// UponLookupBatch registers a handler for the LookupBatch events.
func UponLookupBatch(m dsl.Module, handler func(batchID t.BatchID, origin *batchdbpb.LookupBatchOrigin) error) {
	UponEvent[*batchdbpb.Event_Lookup](m, func(ev *batchdbpb.LookupBatch) error {
		return handler(t.BatchID(ev.BatchId), ev.Origin)
	})
}

// UponLookupBatchResponse registers a handler for the LookupBatchResponse events.
func UponLookupBatchResponse[C any](m dsl.Module, handler func(found bool, txs []*requestpb.Request, metadata []byte, context *C) error) {
	UponEvent[*batchdbpb.Event_LookupResponse](m, func(ev *batchdbpb.LookupBatchResponse) error {
		OriginWrapper, ok := ev.Origin.Type.(*batchdbpb.LookupBatchOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(OriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Found, ev.Txs, ev.Metadata, context)
	})
}

// UponStoreBatch registers a handler for the StoreBatch events.
func UponStoreBatch(m dsl.Module, handler func(batchID t.BatchID, txIDs []t.TxID, txs []*requestpb.Request, metadata []byte, origin *batchdbpb.StoreBatchOrigin) error) {
	UponEvent[*batchdbpb.Event_Store](m, func(ev *batchdbpb.StoreBatch) error {
		return handler(t.BatchID(ev.BatchId), t.TxIDSlice(ev.TxIds), ev.Txs, ev.Metadata, ev.Origin)
	})
}

// UponBatchStored registers a handler for the BatchStored events.
func UponBatchStored[C any](m dsl.Module, handler func(context *C) error) {
	UponEvent[*batchdbpb.Event_Stored](m, func(ev *batchdbpb.BatchStored) error {
		originWrapper, ok := ev.Origin.Type.(*batchdbpb.StoreBatchOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(context)
	})
}
