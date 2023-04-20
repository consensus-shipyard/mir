// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package deploytest

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
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
	_, err := modules.ApplyEventsSequentially(eventList, func(e *eventpb.Event) (*events.EventList, error) {
		return events.EmptyList(), m.applyEvent(ctx, e)
	})
	return err
}

func (m *simTimerModule) applyEvent(ctx context.Context, e *eventpb.Event) error {
	switch e := e.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
	case *eventpb.Event_Timer:
		switch e := e.Timer.Type.(type) {
		case *eventpb.TimerEvent_Delay:
			evtsOut := events.ListOf(e.Delay.EventsToDelay...)
			d := types.Duration(e.Delay.Delay)
			m.delay(ctx, evtsOut, d)
		case *eventpb.TimerEvent_Repeat:
			evtsOut := events.ListOf(e.Repeat.EventsToRepeat...)
			d := types.Duration(e.Repeat.Delay)
			retIdx := tt.RetentionIndex(e.Repeat.RetentionIndex)
			m.repeat(ctx, evtsOut, d, retIdx)
		case *eventpb.TimerEvent_GarbageCollect:
			retIdx := tt.RetentionIndex(e.GarbageCollect.RetentionIndex)
			m.garbageCollect(retIdx)
		default:
			return fmt.Errorf("unexpected type of Timer sub-event: %T", e)
		}
	default:
		return fmt.Errorf("unexpected type of Timer event: %T", e)
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
