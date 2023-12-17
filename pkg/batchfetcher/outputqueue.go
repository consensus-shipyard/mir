package batchfetcher

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/stdtypes"
)

type outputItem struct {
	event stdtypes.Event
	f     func(e stdtypes.Event)
}

type outputQueue struct {
	items []*outputItem
}

func (oq *outputQueue) Enqueue(item *outputItem) {
	oq.items = append(oq.items, item)
}

func (oq *outputQueue) Flush(m dsl.Module) {
	for len(oq.items) > 0 && oq.items[0].event != nil {
		// Convenience variable.
		item := oq.items[0]

		// Execute event output hook.
		if item.f != nil {
			item.f(item.event)
		}

		// Emit queued event.
		dsl.EmitEvent(m, item.event)

		// Remove item from queue.
		oq.items = oq.items[1:]
	}
}
