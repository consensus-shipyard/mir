package stdevents

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type GarbageCollect struct {
	mirEvent
	RetentionIndex stdtypes.RetentionIndex
}

func NewGarbageCollect(dest stdtypes.ModuleID, retIdx stdtypes.RetentionIndex) *GarbageCollect {
	return &GarbageCollect{
		mirEvent: mirEvent{
			DestModule: dest,
		},
		RetentionIndex: retIdx,
	}
}

func NewGarbageCollectWithSrc(
	src stdtypes.ModuleID,
	dest stdtypes.ModuleID,
	retIdx stdtypes.RetentionIndex,
) *GarbageCollect {
	e := NewGarbageCollect(dest, retIdx)
	e.SrcModule = src
	return e
}

func (e *GarbageCollect) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *GarbageCollect) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *GarbageCollect) ToBytes() ([]byte, error) {
	return serialize(e)
}

func (e *GarbageCollect) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}
