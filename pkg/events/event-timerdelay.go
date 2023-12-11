package events

import (
	"fmt"
	"time"

	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	es "github.com/go-errors/errors"
)

type serializableTimerDelay struct {
	mirEvent
	Events []*Raw
	Delay  time.Duration
}

func (std *serializableTimerDelay) TimerDelay() *TimerDelay {
	return &TimerDelay{
		mirEvent: std.mirEvent,
		Events:   sliceutil.Transform(std.Events, func(_ int, raw *Raw) Event { return Event(raw) }),
		Delay:    std.Delay,
	}
}

type TimerDelay struct {
	mirEvent
	Events []Event
	Delay  time.Duration
}

func (e *TimerDelay) serializable() (*serializableTimerDelay, error) {

	// Serialize individual events contained in the TimerDelay event
	// and transform each into a Raw event.
	rawEvents := make([]*Raw, len(e.Events))
	var err error
	for i := range rawEvents {
		rawEvents[i], err = WrapInRaw(e.Events[i])
		if err != nil {
			return nil, es.Errorf("failed serializing delayed event at index %d: %w", i, err)
		}
	}

	return &serializableTimerDelay{
		mirEvent: e.mirEvent,
		Events:   rawEvents,
		Delay:    e.Delay,
	}, nil
}

func NewTimerDelay(dest t.ModuleID, delay time.Duration, events ...Event) *TimerDelay {
	return &TimerDelay{
		mirEvent: mirEvent{DestModule: dest},
		Events:   events,
		Delay:    delay,
	}
}

func NewTimerDelayWithSrc(src t.ModuleID, dest t.ModuleID, delay time.Duration, events ...Event) *TimerDelay {
	e := NewTimerDelay(dest, delay, events...)
	e.SrcModule = src
	return e
}

func (e *TimerDelay) NewSrc(newSrc t.ModuleID) Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *TimerDelay) NewDest(newDest t.ModuleID) Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *TimerDelay) ToBytes() ([]byte, error) {
	serializable, err := e.serializable()
	if err != nil {
		return nil, err
	}

	return serialize(serializable)
}

func (e *TimerDelay) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}
