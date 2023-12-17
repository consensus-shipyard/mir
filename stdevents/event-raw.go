package stdevents

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type Raw struct {
	mirEvent
	Data []byte
}

func NewRaw(dest stdtypes.ModuleID, data []byte) *Raw {
	return &Raw{
		mirEvent: mirEvent{
			DestModule: dest,
		},
		Data: data,
	}
}

func NewRawWithSrc(src stdtypes.ModuleID, dest stdtypes.ModuleID, data []byte) *Raw {
	e := NewRaw(dest, data)
	e.SrcModule = src
	return e
}

func (e *Raw) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *Raw) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
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

func WrapInRaw(e stdtypes.Event) (*Raw, error) {
	data, err := e.ToBytes()
	if err != nil {
		return nil, err
	}
	return NewRawWithSrc(e.Src(), e.Dest(), data), nil
}
