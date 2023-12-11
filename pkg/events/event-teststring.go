package events

import (
	"fmt"

	t "github.com/filecoin-project/mir/pkg/types"
)

type TestString struct {
	mirEvent
	Value string
}

func NewTestString(dest t.ModuleID, value string) *TestString {
	return &TestString{
		mirEvent{
			DestModule: dest,
		},
		value,
	}
}

func NewTestStringWithSrc(src t.ModuleID, dest t.ModuleID, value string) *TestString {
	e := NewTestString(dest, value)
	e.SrcModule = src
	return e
}

func (e *TestString) NewSrc(newSrc t.ModuleID) Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *TestString) NewDest(newDest t.ModuleID) Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *TestString) ToBytes() ([]byte, error) {
	return serialize(e)
}

func (e *TestString) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}
