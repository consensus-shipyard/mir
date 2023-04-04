package events

import (
	bfpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// These are specific events that are not automatically generated but are necessary for the batchfetcher
// to queue them and treat them in order.

// Event creates an eventpb.Event out of an batchfetcherpb.Event.
func Event(dest t.ModuleID, ev *bfpb.Event) *eventpb.Event {
	return &eventpb.Event{
		DestModule: dest.Pb(),
		Type: &eventpb.Event_BatchFetcher{
			BatchFetcher: ev,
		},
	}
}

func ClientProgress(dest t.ModuleID, progress *commonpb.ClientProgress) *eventpb.Event {
	return Event(dest, &bfpb.Event{Type: &bfpb.Event_ClientProgress{ClientProgress: progress}})
}
