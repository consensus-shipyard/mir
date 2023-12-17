// This file extends the protobuf generated Event type by additional methods
// so it can be used as a mir.Event.
// After generating Go code from .proto files, the content of this file is copied (by code in generate.go)
// to eventinterface.go in the folder where the generated Event code is (pkg/pb/eventpb/).
// To modify this file, modify its template in in the protos directory.

package eventpb

import (
	"google.golang.org/protobuf/proto"

	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/stdtypes"
)

// Src returns the module that emitted the event.
// While this information is not always necessary for the system operation,
// it is useful for analyzing event traces and debugging.
func (event *Event) Src() t.ModuleID {
	return ""
}

// NewSrc returns a shallow copy of this Event with an updated source module ID.
// This method is not yet implemented in proto events.
func (event *Event) NewSrc(newSrc t.ModuleID) stdtypes.Event {
	panic("Proto events do not contain source information yet")
}

// Dest returns the destination module of the event.
func (event *Event) Dest() t.ModuleID {
	return t.ModuleID(event.DestModule)
}

// NewDest returns a shallow copy of the event with a new destination module ID.
func (event *Event) NewDest(newDest t.ModuleID) stdtypes.Event {
	newEvent := Event{
		DestModule: newDest.Pb(),
		Type:       event.Type,
	}
	newEvent.DestModule = newDest.Pb()
	return &newEvent
}

// ToBytes returns a serialized representation of the event
// as a slice of bytes from which the event can be reconstructed.
// Note that ToBytes does not necessarily guarantee the output to be deterministic.
// Even multiple subsequent calls to ToBytes on the same event object might return different byte slices.
func (event *Event) ToBytes() ([]byte, error) {
	data, err := proto.Marshal(event)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (event *Event) ToString() string {
	return event.String()
}

// List returns EventList containing the given elements.
func List(evts ...*Event) *stdtypes.EventList {
	res := stdtypes.EmptyList()
	for _, ev := range evts {
		res.PushBack(ev)
	}
	return res
}
