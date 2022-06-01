package timer

import (
	"context"
	"fmt"
	"github.com/filecoin-project/mir/pkg/activemodule"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	t "github.com/filecoin-project/mir/pkg/types"
	"time"
)

// The Timer module abstracts the passage of real time.
// It is used for implementing timeouts and periodic repetition.
type Timer struct {
	retIndex t.TimerRetIndex
}

func (tm *Timer) Run(
	ctx context.Context,
	eventsIn <-chan *events.EventList,
	eventsOut chan<- *events.EventList,
	interceptor eventlog.Interceptor,
) error {
	return activemodule.RunActiveModule(ctx, eventsIn, eventsOut, interceptor, tm.processEventList)
}

func (tm *Timer) Status() (s *statuspb.ProtocolStatus, err error) {
	//TODO implement me
	panic("implement me")
}

func (tm *Timer) processEventList(ctx context.Context, eventList *events.EventList, notifyChan chan<- *events.EventList) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		// Based on event type, invoke the appropriate Timer function.
		// Note that events that later return to the event loop need to be copied in order to prevent a race condition
		// when they are later stripped off their follow-ups, as this happens potentially concurrently
		// with the original event being processed by the interceptor.
		switch e := event.Type.(type) {
		case *eventpb.Event_TimerDelay:
			tm.Delay(
				ctx,
				(&events.EventList{}).PushBackSlice(e.TimerDelay.Events),
				t.TimeDuration(e.TimerDelay.Delay),
				notifyChan,
			)
		case *eventpb.Event_TimerRepeat:
			tm.Repeat(
				ctx,
				(&events.EventList{}).PushBackSlice(e.TimerRepeat.Events),
				t.TimeDuration(e.TimerRepeat.Delay),
				t.TimerRetIndex(e.TimerRepeat.RetentionIndex),
				notifyChan,
			)
		case *eventpb.Event_TimerGarbageCollect:
			tm.GarbageCollect(t.TimerRetIndex(e.TimerGarbageCollect.RetentionIndex))
		default:
			return fmt.Errorf("unexpected type of Timer event: %T", event.Type)
		}
	}

	return nil
}

// Delay writes events to notifyChan after delay.
// If context is canceled, no events might be written.
func (tm *Timer) Delay(
	ctx context.Context,
	events *events.EventList,
	delay t.TimeDuration,
	notifyChan chan<- *events.EventList,
) {

	// This manual implementation has the following advantage over simply using time.AfterFunc:
	// When the context is canceled, the timer stops immediately
	// and does not still fire after a (potentially long) delay.
	go func() {
		select {
		case <-time.After(time.Duration(delay)):
			// If the delay elapses before the context is canceled,
			// try to write events to notifyChan until the context is canceled.
			select {
			case notifyChan <- events:
			case <-ctx.Done():
			}
		case <-ctx.Done():
			// If the context is canceled before the delay elapses, do nothing.
		}
	}()
}

// Repeat periodically writes events to notifyChan with the given period.
// The repeated producing of events is associated with the retention index retIndex.
// The first write of events to notifyChan happens immediately after invocation of Repeat.
// Repeat stops writing events to notifyChan when
// 1) ctx is canceled or
// 2) GarbageCollect is called with an argument greater than the associated retIndex.
func (tm *Timer) Repeat(
	ctx context.Context,
	events *events.EventList,
	period t.TimeDuration,
	retIndex t.TimerRetIndex,
	notifyChan chan<- *events.EventList,
) {
	go func() {

		// Create a ticker with the given period.
		ticker := time.NewTicker(time.Duration(period))
		defer ticker.Stop()

		// Repeat as long as this repetition task has not been garbage-collected.
		for retIndex >= tm.retIndex {

			// Try to write events to notifyChan until the context is canceled.
			select {
			case notifyChan <- events:
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
func (tm *Timer) GarbageCollect(retIndex t.TimerRetIndex) {

	// Retention index must be monotonically increasing over calls to GarbageCollect.
	if retIndex > tm.retIndex {
		tm.retIndex = retIndex
	}
}
