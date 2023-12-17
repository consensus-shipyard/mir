package internal

import (
	"github.com/filecoin-project/mir/stdtypes"
)

type ModuleImpl interface {
	Event(ev stdtypes.Event) (*stdtypes.EventList, error)
}
