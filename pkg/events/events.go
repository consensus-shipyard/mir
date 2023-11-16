package events

import (
	"encoding/json"

	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// Event Constructors
// ============================================================
// TODO: Put these events in a separate package.

type TestStringEvent struct {
	SrcModule  t.ModuleID
	DestModule t.ModuleID
	Value      string
}

func (t *TestStringEvent) Src() t.ModuleID {
	return t.SrcModule
}

func (t *TestStringEvent) NewSrc(newSrc t.ModuleID) Event {
	newTestString := *t
	newTestString.SrcModule = newSrc
	return &newTestString
}

func (t *TestStringEvent) Dest() t.ModuleID {
	return t.DestModule
}

func (t *TestStringEvent) NewDest(newDest t.ModuleID) Event {
	newTestString := *t
	newTestString.DestModule = newDest
	return &newTestString
}

func (t *TestStringEvent) ToBytes() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TestStringEvent) ToString() string {
	bytes, err := t.ToBytes()
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

type TestUintEvent struct {
	SrcModule  t.ModuleID
	DestModule t.ModuleID
	Value      uint64
}

func (t *TestUintEvent) Src() t.ModuleID {
	return t.SrcModule
}

func (t *TestUintEvent) NewSrc(newSrc t.ModuleID) Event {
	newTestUint := *t
	newTestUint.SrcModule = newSrc
	return &newTestUint
}

func (t *TestUintEvent) Dest() t.ModuleID {
	return t.DestModule
}

func (t *TestUintEvent) NewDest(newDest t.ModuleID) Event {
	newTestUint := *t
	newTestUint.DestModule = newDest
	return &newTestUint
}

func (t *TestUintEvent) ToBytes() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TestUintEvent) ToString() string {
	bytes, err := t.ToBytes()
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func NewTestString(dest t.ModuleID, s string) Event {
	return &TestStringEvent{
		SrcModule:  "",
		DestModule: dest,
		Value:      s,
	}
}

func NewTestUint(dest t.ModuleID, u uint64) Event {
	return &TestUintEvent{
		SrcModule:  "",
		DestModule: dest,
		Value:      u,
	}
}
