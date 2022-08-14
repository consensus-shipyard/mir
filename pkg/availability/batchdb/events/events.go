package events

import (
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Event creates an eventpb.Event out of a batchdbpb.Event.
func Event(dest t.ModuleID, ev *batchdbpb.Event) *eventpb.Event {
	return &eventpb.Event{
		DestModule: dest.Pb(),
		Type: &eventpb.Event_BatchDb{
			BatchDb: ev,
		},
	}
}

// LookupBatch is used to pull a batch with its metadata from the local batch database.
func LookupBatch(dest t.ModuleID, batchID t.BatchID, origin *batchdbpb.LookupBatchOrigin) *eventpb.Event {
	return Event(dest, &batchdbpb.Event{
		Type: &batchdbpb.Event_Lookup{
			Lookup: &batchdbpb.LookupBatch{
				BatchId: batchID.Pb(),
				Origin:  origin,
			},
		},
	})
}

// LookupBatchResponse is a response to a LookupBatch event.
func LookupBatchResponse(dest t.ModuleID, found bool, txs []*requestpb.Request, metadata []byte, origin *batchdbpb.LookupBatchOrigin) *eventpb.Event {
	return Event(dest, &batchdbpb.Event{
		Type: &batchdbpb.Event_LookupResponse{
			LookupResponse: &batchdbpb.LookupBatchResponse{
				Found:    found,
				Txs:      txs,
				Metadata: metadata,
				Origin:   origin,
			},
		},
	})
}

// StoreBatch is used to store a new batch in the local batch database.
func StoreBatch(dest t.ModuleID, batchID t.BatchID, txIDs []t.TxID, txs []*requestpb.Request, metadata []byte, origin *batchdbpb.StoreBatchOrigin) *eventpb.Event {
	return Event(dest, &batchdbpb.Event{
		Type: &batchdbpb.Event_Store{
			Store: &batchdbpb.StoreBatch{
				BatchId:  batchID.Pb(),
				TxIds:    t.TxIDSlicePb(txIDs),
				Txs:      txs,
				Metadata: metadata,
				Origin:   origin,
			},
		},
	})
}

// BatchStored is a response to a StoreBatch event.
func BatchStored(dest t.ModuleID, origin *batchdbpb.StoreBatchOrigin) *eventpb.Event {
	return Event(dest, &batchdbpb.Event{
		Type: &batchdbpb.Event_Stored{
			Stored: &batchdbpb.BatchStored{
				Origin: origin,
			},
		},
	})
}
