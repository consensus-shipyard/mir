package events

import (
	bfpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Event creates an eventpb.Event out of an batchfetcherpb.Event.
func Event(dest t.ModuleID, ev *bfpb.Event) *eventpb.Event {
	return &eventpb.Event{
		DestModule: dest.Pb(),
		Type: &eventpb.Event_BatchFetcher{
			BatchFetcher: ev,
		},
	}
}

// NewOrderedBatch notifies the application when an ordered batch containing transaction payloads
// has been fetched from the availability layer and is ready for execution.
func NewOrderedBatch(dest t.ModuleID, txs []*requestpb.Request) *eventpb.Event {
	return Event(dest, &bfpb.Event{
		Type: &bfpb.Event_NewOrderedBatch{
			NewOrderedBatch: &bfpb.NewOrderedBatch{
				Txs: txs,
			},
		},
	})
}
