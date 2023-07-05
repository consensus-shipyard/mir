package timer

import (
	"context"
	"sync"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/events"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

// The Timer module abstracts the passage of real time.
// It is used for implementing timeouts and periodic repetition.
type Timer struct {
	eventsOut chan *events.EventList

	retIndexMutex sync.RWMutex
	retIndex      tt.RetentionIndex
}

func New() *Timer {
	return &Timer{
		eventsOut: make(chan *events.EventList),
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (tm *Timer) ImplementsModule() {}

func (tm *Timer) EventsOut() <-chan *events.EventList {
	return tm.eventsOut
}

func (tm *Timer) ApplyEvents(ctx context.Context, eventList *events.EventList) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		// Based on event type, invoke the appropriate Timer function.
		// Note that events that later return to the event loop need to be copied in order to prevent a race condition
		// when they are later stripped off their follow-ups, as this happens potentially concurrently
		// with the original event being processed by the interceptor.
		switch e := event.Type.(type) {
		case *eventpbtypes.Event_Init:
			// no actions on init
		case *eventpbtypes.Event_Timer:
			switch e := e.Timer.Type.(type) {
			case *eventpbtypes.TimerEvent_Delay:
				tm.Delay(
					ctx,
					events.ListOf(e.Delay.EventsToDelay...),
					e.Delay.Delay,
				)
			case *eventpbtypes.TimerEvent_Repeat:
				tm.Repeat(
					ctx,
					events.ListOf(e.Repeat.EventsToRepeat...),
					e.Repeat.Delay,
					e.Repeat.RetentionIndex,
				)
			case *eventpbtypes.TimerEvent_GarbageCollect:
				tm.GarbageCollect(e.GarbageCollect.RetentionIndex)
			}
		default:
			return es.Errorf("unexpected type of Timer event: %T", event.Type)
		}
	}

	return nil
}

// Delay writes events to the output channel after delay.
// If context is canceled, no events might be written.
func (tm *Timer) Delay(
	ctx context.Context,
	events *events.EventList,
	delay types.Duration,
) {

	// This manual implementation has the following advantage over simply using time.AfterFunc:
	// When the context is canceled, the timer stops immediately
	// and does not still fire after a (potentially long) delay.
	go func() {
		select {
		case <-time.After(time.Duration(delay)):
			// If the delay elapses before the context is canceled,
			// try to write events to the output channel until the context is canceled.
			select {
			case tm.eventsOut <- events:
			case <-ctx.Done():
			}
		case <-ctx.Done():
			// If the context is canceled before the delay elapses, do nothing.
		}
	}()
}

// Repeat periodically writes events to the output channel with the given period.
// The repeated producing of events is associated with the retention index retIndex.
// The first write of events to the output channel happens immediately after invocation of Repeat.
// Repeat stops writing events to the output channel when
// 1) ctx is canceled or
// 2) GarbageCollect is called with an argument greater than the associated retIndex.
func (tm *Timer) Repeat(
	ctx context.Context,
	events *events.EventList,
	period types.Duration,
	retIndex tt.RetentionIndex,
) {
	go func() {

		// Create a ticker with the given period.
		ticker := time.NewTicker(time.Duration(period))
		defer ticker.Stop()

		// Repeat as long as this repetition task has not been garbage-collected.
		for retIndex >= tm.getRetIndex() {

			// Try to write events to the output channel until the context is canceled.
			select {
			case tm.eventsOut <- events:
			case <-ctx.Done():
				// If the context is canceled before the events can be
				// written to the output channel, stop the repetition.
				return
			}

			// Wait for the defined period of time (or until the context is canceled)
			select {
			case <-ticker.C:
			case <-ctx.Done():
				// If the context is canceled before the delay elapses, stop the repetition.
				return
			}
		}

	}()
}

// GarbageCollect cancels the effect of outdated calls to Repeat.
// When GarbageCollect is called, the Timer stops repeatedly producing events associated with a retention index
// smaller than retIndex.
// If GarbageCollect already has been invoked with the same or higher retention index, the call has no effect.
func (tm *Timer) GarbageCollect(retIndex tt.RetentionIndex) {
	tm.retIndexMutex.Lock()
	defer tm.retIndexMutex.Unlock()

	// Retention index must be monotonically increasing over calls to GarbageCollect.
	if retIndex > tm.retIndex {
		tm.retIndex = retIndex
	}
}

func (tm *Timer) getRetIndex() tt.RetentionIndex {
	tm.retIndexMutex.RLock()
	defer tm.retIndexMutex.RUnlock()

	return tm.retIndex
}
