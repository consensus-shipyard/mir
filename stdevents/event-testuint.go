package stdevents

import (
	"fmt"

	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/stdtypes"
)

type TestUint64 struct {
	mirEvent
	Value uint64
}

func NewTestUint64(dest t.ModuleID, value uint64) *TestUint64 {
	return &TestUint64{
		mirEvent{
			DestModule: dest,
		},
		value,
	}
}

func NewTestUint64WithSrc(src t.ModuleID, dest t.ModuleID, value uint64) *TestUint64 {
	e := NewTestUint64(dest, value)
	e.SrcModule = src
	return e
}

func (e *TestUint64) NewSrc(newSrc t.ModuleID) stdtypes.Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *TestUint64) NewDest(newDest t.ModuleID) stdtypes.Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *TestUint64) ToBytes() ([]byte, error) {
	return serialize(e)
}

func (e *TestUint64) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}
