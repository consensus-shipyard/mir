package modules

import (
	"context"
	"github.com/filecoin-project/mir/pkg/events"
	t "github.com/filecoin-project/mir/pkg/types"
)

// The Timer module abstracts the passage of real time.
// It is used for implementing timeouts and periodic repetition.
type Timer interface {

	// Delay writes events to notifyChan after delay.
	// If context is canceled, no events might be written.
	Delay(ctx context.Context, events *events.EventList, delay t.TimeDuration, notifyChan chan<- *events.EventList)

	// Repeat periodically writes events to notifyChan with the given period.
	// The repeated producing of events is associated with the retention index retIndex.
	// The first write of events to notifyChan happens immediately after invocation of Repeat.
	// Repeat stops writing events to notifyChan when
	// 1) ctx is canceled or
	// 2) GarbageCollect is called with an argument greater than the associated retIndex.
	// If GarbageCollect has already been invoked with an argument higher than retIndex, Repeat has no effect.
	Repeat(
		ctx context.Context,
		events *events.EventList,
		period t.TimeDuration,
		retIndex t.TimerRetIndex,
		notifyChan chan<- *events.EventList,
	)

	// GarbageCollect cancels the effect of outdated calls to Repeat.
	// When GarbageCollect is called, the Timer stops repeatedly producing events associated with a retention index
	// smaller than retIndex.
	// The timer may ignore the call to GarbageCollect
	// if GarbageCollect already has been invoked with the same or higher retention index.
	GarbageCollect(retIndex t.TimerRetIndex)
}
