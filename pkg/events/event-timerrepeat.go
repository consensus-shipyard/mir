package events

import (
	"fmt"
	"time"

	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	es "github.com/go-errors/errors"
)

type serializableTimerRepeat struct {
	mirEvent
	Events         []*Raw
	Period         time.Duration
	RetentionIndex RetentionIndex
}

func (str *serializableTimerRepeat) TimerRepeat() *TimerRepeat {
	return &TimerRepeat{
		mirEvent:       str.mirEvent,
		Events:         sliceutil.Transform(str.Events, func(_ int, raw *Raw) Event { return Event(raw) }),
		Period:         str.Period,
		RetentionIndex: str.RetentionIndex,
	}
}

type TimerRepeat struct {
	mirEvent
	Events         []Event
	Period         time.Duration
	RetentionIndex RetentionIndex
}

func (e *TimerRepeat) serializable() (*serializableTimerRepeat, error) {
	// Serialize individual events contained in the TimerDelay event
	// and transform each into a Raw event.
	rawEvents := make([]*Raw, len(e.Events))
	var err error
	for i := range rawEvents {
		rawEvents[i], err = WrapInRaw(e.Events[i])
		if err != nil {
			return nil, es.Errorf("failed serializing repeated event at index %d: %w", i, err)
		}
	}

	return &serializableTimerRepeat{
		mirEvent:       e.mirEvent,
		Events:         rawEvents,
		Period:         e.Period,
		RetentionIndex: e.RetentionIndex,
	}, nil
}

func NewTimerRepeat(
	dest t.ModuleID,
	period time.Duration,
	retentionIndex RetentionIndex,
	events ...Event,
) *TimerRepeat {
	return &TimerRepeat{
		mirEvent:       mirEvent{DestModule: dest},
		Events:         events,
		Period:         period,
		RetentionIndex: retentionIndex,
	}
}

func NewTimerRepeatWithSrc(
	src t.ModuleID,
	dest t.ModuleID,
	period time.Duration,
	retentionIndex RetentionIndex,
	events ...Event,
) *TimerRepeat {
	e := NewTimerRepeat(dest, period, retentionIndex, events...)
	e.SrcModule = src
	return e
}

func (e *TimerRepeat) NewSrc(newSrc t.ModuleID) Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *TimerRepeat) NewDest(newDest t.ModuleID) Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *TimerRepeat) ToBytes() ([]byte, error) {
	serializable, err := e.serializable()
	if err != nil {
		return nil, err
	}

	return serialize(serializable)
}

func (e *TimerRepeat) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}
