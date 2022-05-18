package timer

import (
	"context"
	"github.com/filecoin-project/mir/pkg/events"
	t "github.com/filecoin-project/mir/pkg/types"
	"time"
)

// The Timer module abstracts the passage of real time.
// It is used for implementing timeouts and periodic repetition.
type Timer struct {
	retIndex t.TimerRetIndex
}

// Delay writes events to notifyChan after delay.
// If context is canceled, no events might be written.
func (t *Timer) Delay(
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
func (t *Timer) Repeat(
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
		for retIndex >= t.retIndex {

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
func (t *Timer) GarbageCollect(retIndex t.TimerRetIndex) {

	// Retention index must be monotonically increasing over calls to GarbageCollect.
	if retIndex > t.retIndex {
		t.retIndex = retIndex
	}
}
