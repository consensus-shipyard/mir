// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package deploytest

import (
	"context"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/testsim"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

type simTimerModule struct {
	*SimNode
	eventsOut chan *stdtypes.EventList
	processes map[tt.RetentionIndex]*testsim.Process
}

// NewSimTimerModule returns a Timer modules to be used in simulation.
func NewSimTimerModule(node *SimNode) modules.ActiveModule {
	return &simTimerModule{
		SimNode:   node,
		eventsOut: make(chan *stdtypes.EventList, 1),
		processes: map[tt.RetentionIndex]*testsim.Process{},
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

	// We only support proto events.
	pbevent, ok := event.(*eventpb.Event)
	if !ok {
		return es.Errorf("The simulation timer module only supports proto events, received %T", event)
	}

	switch e := pbevent.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
	case *eventpb.Event_Timer:
		switch e := e.Timer.Type.(type) {
		case *eventpb.TimerEvent_Delay:
			evtsOut := eventpb.List(e.Delay.EventsToDelay...)
			d := types.Duration(e.Delay.Delay)
			m.delay(ctx, evtsOut, d)
		case *eventpb.TimerEvent_Repeat:
			evtsOut := eventpb.List(e.Repeat.EventsToRepeat...)
			d := types.Duration(e.Repeat.Delay)
			retIdx := tt.RetentionIndex(e.Repeat.RetentionIndex)
			m.repeat(ctx, evtsOut, d, retIdx)
		case *eventpb.TimerEvent_GarbageCollect:
			retIdx := tt.RetentionIndex(e.GarbageCollect.RetentionIndex)
			m.garbageCollect(retIdx)
		default:
			return es.Errorf("unexpected type of Timer sub-event: %T", e)
		}
	default:
		return es.Errorf("unexpected type of Timer event: %T", e)
	}

	return nil
}

func (m *simTimerModule) delay(ctx context.Context, eventList *stdtypes.EventList, d types.Duration) {
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

func (m *simTimerModule) repeat(ctx context.Context, eventList *stdtypes.EventList, d types.Duration, retIdx tt.RetentionIndex) {
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
