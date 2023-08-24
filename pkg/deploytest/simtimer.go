// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package deploytest

import (
	"context"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	"github.com/filecoin-project/mir/pkg/testsim"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

type simTimerModule struct {
	*SimNode
	eventsOut chan *events.EventList
	processes map[tt.RetentionIndex]*testsim.Process
}

// NewSimTimerModule returns a Timer modules to be used in simulation.
func NewSimTimerModule(node *SimNode) modules.ActiveModule {
	return &simTimerModule{
		SimNode:   node,
		eventsOut: make(chan *events.EventList, 1),
		processes: map[tt.RetentionIndex]*testsim.Process{},
	}
}

func (m *simTimerModule) ImplementsModule() {}

func (m *simTimerModule) EventsOut() <-chan *events.EventList {
	return m.eventsOut
}

func (m *simTimerModule) ApplyEvents(ctx context.Context, eventList *events.EventList) error {
	_, err := modules.ApplyEventsSequentially(eventList, func(e *eventpbtypes.Event) (*events.EventList, error) {
		return events.EmptyList(), m.applyEvent(ctx, e)
	})
	return err
}

func (m *simTimerModule) applyEvent(ctx context.Context, e *eventpbtypes.Event) error {
	switch e := e.Type.(type) {
	case *eventpbtypes.Event_Init:
		// no actions on init
	case *eventpbtypes.Event_Timer:
		switch e := e.Timer.Type.(type) {
		case *eventpbtypes.TimerEvent_Delay:
			evtsOut := events.ListOf(e.Delay.EventsToDelay...)
			m.delay(ctx, evtsOut, e.Delay.Delay)
		case *eventpbtypes.TimerEvent_Repeat:
			evtsOut := events.ListOf(e.Repeat.EventsToRepeat...)
			m.repeat(ctx, evtsOut, e.Repeat.Delay, e.Repeat.RetentionIndex)
		case *eventpbtypes.TimerEvent_GarbageCollect:
			m.garbageCollect(e.GarbageCollect.RetentionIndex)
		default:
			return es.Errorf("unexpected type of Timer sub-event: %T", e)
		}
	default:
		return es.Errorf("unexpected type of Timer event: %T", e)
	}

	return nil
}

func (m *simTimerModule) delay(ctx context.Context, eventList *events.EventList, d types.Duration) {
	proc := m.Spawn()

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			proc.Kill()
		case <-done:
		}
	}()

	go func() {
		defer close(done)

		if !proc.Delay(time.Duration(d)) {
			return
		}

		select {
		case eventsOut := <-m.eventsOut:
			eventsOut.PushBackList(eventList)
			m.eventsOut <- eventsOut
		default:
			m.eventsOut <- eventList
		}
		m.SimNode.SendEvents(proc, eventList)

		proc.Exit()
	}()
}

func (m *simTimerModule) repeat(ctx context.Context, eventList *events.EventList, d types.Duration, retIdx tt.RetentionIndex) {
	proc := m.Spawn()
	m.processes[retIdx] = proc

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			proc.Kill()
		case <-done:
		}
	}()

	go func() {
		defer close(done)

		for {
			if !proc.Delay(time.Duration(d)) {
				return
			}

			select {
			case eventsOut := <-m.eventsOut:
				eventsOut.PushBackList(eventList)
				m.eventsOut <- eventsOut
			default:
				m.eventsOut <- eventList
			}
			m.SimNode.SendEvents(proc, eventList)
		}
	}()
}

func (m *simTimerModule) garbageCollect(retIdx tt.RetentionIndex) {
	for i, proc := range m.processes {
		if i < retIdx {
			proc.Kill()
			delete(m.processes, i)
		}
	}
}
