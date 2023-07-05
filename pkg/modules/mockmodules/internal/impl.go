package internal

import (
	"github.com/filecoin-project/mir/pkg/events"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
)

type ModuleImpl interface {
	Event(ev *eventpbtypes.Event) (*events.EventList, error)
}
