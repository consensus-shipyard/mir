// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package deploytest

import (
	"context"
	"time"

	"github.com/filecoin-project/mir/stdevents"
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/testsim"
)

type simTimerModule struct {
	*SimNode
	eventsOut chan *stdtypes.EventList
	processes map[stdtypes.RetentionIndex]*testsim.Process
}

// NewSimTimerModule returns a Timer modules to be used in simulation.
func NewSimTimerModule(node *SimNode) modules.ActiveModule {
	return &simTimerModule{
		SimNode:   node,
		eventsOut: make(chan *stdtypes.EventList, 1),
		processes: map[stdtypes.RetentionIndex]*testsim.Process{},
	}
}

func (m *simTimerModule) ImplementsModule() {}

func (m *simTimerModule) EventsOut() <-chan *stdtypes.EventList {
	return m.eventsOut
}

func (m *simTimerModule) ApplyEvents(ctx context.Context, eventList *stdtypes.EventList) error {
	_, err := modules.ApplyEventsSequentially(eventList, func(e stdtypes.Event) (*stdtypes.EventList, error) {
		return stdtypes.EmptyList(), m.applyEvent(ctx, e)
	})
	return err
}

func (m *simTimerModule) applyEvent(ctx context.Context, event stdtypes.Event) error {

	switch evt := event.(type) {
	case *eventpb.Event:
		switch e := evt.Type.(type) {
		case *eventpb.Event_Init:
			// no actions on init
		default:
			return es.Errorf("unexpected type of Timer event: %T", e)
		}
	case *stdevents.TimerDelay:
		evtsOut := stdtypes.ListOf(evt.Events...)
		m.delay(ctx, evtsOut, evt.Delay)
	case *stdevents.TimerRepeat:
		evtsOut := stdtypes.ListOf(evt.Events...)
		m.repeat(ctx, evtsOut, evt.Period, evt.RetentionIndex)
	case *stdevents.GarbageCollect:
		m.garbageCollect(evt.RetentionIndex)
	default:
		return es.Errorf("Unsupported simulation event type: %T", event)
	}

	return nil
}

func (m *simTimerModule) delay(ctx context.Context, eventList *stdtypes.EventList, d time.Duration) {
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

		if !proc.Delay(d) {
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

func (m *simTimerModule) repeat(ctx context.Context, eventList *stdtypes.EventList, d time.Duration, retIdx stdtypes.RetentionIndex) {
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

func (m *simTimerModule) garbageCollect(retIdx stdtypes.RetentionIndex) {
	for i, proc := range m.processes {
		if i < retIdx {
			proc.Kill()
			delete(m.processes, i)
		}
	}
}
