package stdevents

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type Init struct {
	mirEvent
}

func NewInit(dest stdtypes.ModuleID) *Init {
	return &Init{
		mirEvent{
			DestModule: dest,
		},
	}
}

func NewInitWithSrc(src stdtypes.ModuleID, dest stdtypes.ModuleID) *Init {
	e := NewInit(dest)
	e.SrcModule = src
	return e
}

func (e *Init) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *Init) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *Init) ToBytes() ([]byte, error) {
	return serialize(e)
}

func (e *Init) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}
