//nolint:dupl
package stdevents

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

// TODO: Make this generic in the form of TestValue[T any] (or some subset of any)
//   Then the linter will stop complaining about duplication.

type TestString struct {
	mirEvent
	Value string
}

func NewTestString(dest stdtypes.ModuleID, value string) *TestString {
	return &TestString{
		mirEvent{
			DestModule: dest,
		},
		value,
	}
}

func NewTestStringWithSrc(src stdtypes.ModuleID, dest stdtypes.ModuleID, value string) *TestString {
	e := NewTestString(dest, value)
	e.SrcModule = src
	return e
}

func (e *TestString) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *TestString) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
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
