package timer

import (
	"context"
	"sync"
	"time"

	"github.com/filecoin-project/mir/stdevents"
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/stdtypes"
)

// The Timer module abstracts the passage of real time.
// It is used for implementing timeouts and periodic repetition.
type Timer struct {
	eventsOut chan *stdtypes.EventList

	retIndexMutex sync.RWMutex
	retIndex      stdtypes.RetentionIndex
}

func New() *Timer {
	return &Timer{
		eventsOut: make(chan *stdtypes.EventList),
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (tm *Timer) ImplementsModule() {}

func (tm *Timer) EventsOut() <-chan *stdtypes.EventList {
	return tm.eventsOut
}

func (tm *Timer) ApplyEvents(ctx context.Context, eventList *stdtypes.EventList) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		if err := tm.applyEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (tm *Timer) applyEvent(ctx context.Context, event stdtypes.Event) error {

	// Based on event type, invoke the appropriate Timer function.
	// Note that events that later return to the event loop need to be copied in order to prevent a race condition
	// when they are later, as this happens potentially concurrently
	// with the original event being processed by the interceptor.
	switch e := event.(type) {
	case *eventpb.Event:
		// Support for legacy proto events. TODO: Remove this eventually.
		if err := tm.applyLegacyPbEvent(e); err != nil {
			return err
		}
	case *stdevents.Init:
		// Nothing to initialize.
	case *stdevents.TimerDelay:
		tm.Delay(ctx, stdtypes.ListOf(e.Events...), e.Delay)
	case *stdevents.TimerRepeat:
		tm.Repeat(ctx, stdtypes.ListOf(e.Events...), e.Period, e.RetentionIndex)
	case *stdevents.GarbageCollect:
		tm.GarbageCollect(e.RetentionIndex)
	default:
		return es.Errorf("unexpected event type: %T", e)
	}

	return nil
}

// TODO: Remove this function when other modules are updated to use stdevents with the Timer.
func (tm *Timer) applyLegacyPbEvent(event *eventpb.Event) error {
	switch event.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
	default:
		return es.Errorf("unexpected type of Timer event: %T", event.Type)
	}
	return nil
}

// Delay writes events to the output channel after delay.
// If context is canceled, no events might be written.
func (tm *Timer) Delay(
	ctx context.Context,
	events *stdtypes.EventList,
	delay time.Duration,
) {

	// This manual implementation has the following advantage over simply using time.AfterFunc:
	// When the context is canceled, the timer stops immediately
	// and does not still fire after a (potentially long) delay.
	go func() {
		select {
		case <-time.After(delay):
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
	events *stdtypes.EventList,
	period time.Duration,
	retIndex stdtypes.RetentionIndex,
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
func (tm *Timer) GarbageCollect(retIndex stdtypes.RetentionIndex) {
	tm.retIndexMutex.Lock()
	defer tm.retIndexMutex.Unlock()

	// Retention index must be monotonically increasing over calls to GarbageCollect.
	if retIndex > tm.retIndex {
		tm.retIndex = retIndex
	}
}

func (tm *Timer) getRetIndex() stdtypes.RetentionIndex {
	tm.retIndexMutex.RLock()
	defer tm.retIndexMutex.RUnlock()

	return tm.retIndex
}
