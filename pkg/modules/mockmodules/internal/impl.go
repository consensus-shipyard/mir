package internal

import (
	"github.com/filecoin-project/mir/pkg/events"
)

type ModuleImpl interface {
	Event(ev events.Event) (*events.EventList, error)
}
