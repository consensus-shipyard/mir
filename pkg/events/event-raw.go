package events

import (
	"fmt"

	t "github.com/filecoin-project/mir/pkg/types"
)

type Raw struct {
	mirEvent
	Data []byte
}

func NewRaw(dest t.ModuleID, data []byte) *Raw {
	return &Raw{
		mirEvent: mirEvent{
			DestModule: dest,
		},
		Data: data,
	}
}

func NewRawWithSrc(src t.ModuleID, dest t.ModuleID, data []byte) *Raw {
	e := NewRaw(dest, data)
	e.SrcModule = src
	return e
}

func (e *Raw) NewSrc(newSrc t.ModuleID) Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *Raw) NewDest(newDest t.ModuleID) Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *Raw) ToBytes() ([]byte, error) {
	return serialize(e)
}

func (e *Raw) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}

func WrapInRaw(e Event) (*Raw, error) {
	data, err := e.ToBytes()
	if err != nil {
		return nil, err
	}
	return NewRawWithSrc(e.Src(), e.Dest(), data), nil
}
