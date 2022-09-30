package mir

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/modules/mockmodules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

func TestNode_Run(t *testing.T) {
	testCases := map[string]func(t *testing.T) (m modules.Modules, done <-chan struct{}){
		"InitEvents": func(t *testing.T) (modules.Modules, <-chan struct{}) {
			ctrl := gomock.NewController(t)
			mockModule1 := mockmodules.NewMockPassiveModule(ctrl)
			mockModule2 := mockmodules.NewMockPassiveModule(ctrl)

			var wg sync.WaitGroup
			wg.Add(2)

			mockModule1.EXPECT().Event(events.Init("mock1")).
				Do(func(_ any) { wg.Done() }).
				Return(events.EmptyList(), nil)
			mockModule2.EXPECT().Event(events.Init("mock2")).
				Do(func(_ any) { wg.Done() }).
				Return(events.EmptyList(), nil)

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			m := map[types.ModuleID]modules.Module{
				"mock1": mockModule1,
				"mock2": mockModule2,
			}
			return m, done
		},
	}

	for testName, tc := range testCases {
		tc := tc
		t.Run(testName, func(t *testing.T) {
			m, tcDone := tc(t)

			logger := logging.ConsoleWarnLogger
			n, err := NewNode(
				"testnode",
				DefaultNodeConfig().WithLogger(logger),
				m,
				nil,
				nil,
			)

			assert.Nil(t, err)
			ctx, stopNode := context.WithCancel(context.Background())

			nodeStopped := make(chan struct{})
			go func() {
				err := n.Run(ctx)
				assert.Equal(t, ErrStopped, err)
				close(nodeStopped)
			}()

			// Wait until either the test case is done or a 2 seconds deadline
			select {
			case <-tcDone:
			case <-time.After(2 * time.Second):
			}

			stopNode()
			<-nodeStopped
		})
	}
}

func TestNode_Backpressure(t *testing.T) {

	nodeConfig := DefaultNodeConfig()

	// Set an input event rate that would fill the node's event buffers in one second in 10 batches.
	blabberModule := newBlabber(uint64(nodeConfig.PauseInputThreshold/10), 100*time.Millisecond)

	// Set the event consumption rate to 1/2 of the input rate (i.e., draining the buffer in 2 seconds)
	// and create the consumer module.
	consumerDelay := 2 * time.Second / time.Duration(nodeConfig.PauseInputThreshold)
	consumerModule := newConsumer(consumerDelay)

	// Instantiate node with a fast blabber module and a slow consumer module.
	n, err := NewNode(
		"testnode",
		DefaultNodeConfig(),
		map[types.ModuleID]modules.Module{
			"blabber":  blabberModule,
			"consumer": consumerModule,
		},
		nil,
		nil,
	)
	assert.Nil(t, err)

	// Start a flood of dummy events.
	blabberModule.Go()

	// Start the node.
	nodeError := make(chan error)
	go func() {
		nodeError <- n.Run(context.Background())
	}()

	// Run for 5 seconds, then stop the node.
	time.Sleep(5 * time.Second)
	n.Stop()
	err = <-nodeError
	fmt.Printf("node error: %v\n", err)
	assert.True(t, errors.Is(err, ErrStopped), "unexpected node error: \"%v\", expected \"%v\"", err, ErrStopped)

	// The number of submitted events must not exceed the number of consumed events by too much
	// (accounting for events still in the buffer and the overshooting caused by batched adding of events).
	fmt.Printf("Total submitted events: %d\n", atomic.LoadUint64(&blabberModule.totalSubmitted))
	totalSubmitted := atomic.LoadUint64(&blabberModule.totalSubmitted)
	expectSubmitted := atomic.LoadUint64(&consumerModule.numProcessed) +
		uint64(nodeConfig.PauseInputThreshold) + // Events left in the buffer
		uint64(nodeConfig.MaxEventBatchSize) + // Events in the consumer's processing queue
		2*blabberModule.batchSize // one batch of overshooting, one batch waiting in the babbler's output channel.
	assert.LessOrEqual(t, totalSubmitted, expectSubmitted, "too many events submitted (node event buffer overflow)")

}

// =================================================================================================
// Blabber module (for testing event backpressure)
// =================================================================================================

// The babbler is a simple ActiveModule that just produces batches of dummy events at a given rate.
type blabber struct {
	flood          chan *events.EventList // Output channel for batches of dummy events.
	batchSize      uint64                 // Number of events output at once.
	period         time.Duration          // Time between batches.
	totalSubmitted uint64                 // Counter for total number of submitted events.
}

func newBlabber(batchSize uint64, period time.Duration) *blabber {
	return &blabber{
		flood:     make(chan *events.EventList),
		batchSize: batchSize,
		period:    period,
	}
}

// Go starts babbling, i.e., producing batches of dummy events at the rate configured at instantiation.
// Once started, the babbler is unstoppable and keeps babbling forever.
func (b *blabber) Go() {
	go func() {
		for {
			evts := events.ListOf(sliceutil.Repeat(events.TestingUint("consumer", 0), int(b.batchSize))...)
			b.flood <- evts
			atomic.AddUint64(&b.totalSubmitted, b.batchSize)
			time.Sleep(b.period)
		}
	}()
}

func (b *blabber) ImplementsModule() {}

func (b *blabber) ApplyEvents(_ context.Context, _ *events.EventList) error {
	return nil
}

func (b *blabber) EventsOut() <-chan *events.EventList {
	return b.flood
}

// =================================================================================================
// Consumer module (for testing event backpressure)
// =================================================================================================

// The consumer is a simple passive module that consumes events at a given rate.
// It treats each event as a noop and just counts the number of events it has processed.
type consumer struct {
	delay        time.Duration
	numProcessed uint64
}

func newConsumer(delay time.Duration) *consumer {
	return &consumer{delay: delay}
}

func (c *consumer) ImplementsModule() {}

// ApplyEvents increments a counter and sleeps for a given duration (set at module instantiation)
// for each event in the given list.
func (c *consumer) ApplyEvents(evts *events.EventList) (*events.EventList, error) {
	evtsOut, err := modules.ApplyEventsSequentially(evts, func(event *eventpb.Event) (*events.EventList, error) {
		atomic.AddUint64(&c.numProcessed, 1)
		time.Sleep(c.delay)
		return events.EmptyList(), nil
	})
	return evtsOut, err
}
