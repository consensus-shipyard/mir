package batchfetcher

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

type outputItem struct {
	event *eventpb.Event
}

type outputQueue struct {
	items []*outputItem
}

func (oq *outputQueue) Enqueue(item *outputItem) {
	oq.items = append(oq.items, item)
}

func (oq *outputQueue) Flush(m dsl.Module) {
	for len(oq.items) > 0 && oq.items[0].event != nil {
		dsl.EmitEvent(m, oq.items[0].event)
		oq.items = oq.items[1:]
	}
}
