package internal

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

type ModuleImpl interface {
	Event(ev *eventpb.Event) (*events.EventList, error)
}
