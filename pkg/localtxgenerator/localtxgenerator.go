package localtxgenerator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/logging"
	"go.uber.org/atomic"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	mempoolpbevents "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type ModuleConfig struct {
	Mempool t.ModuleID
}

func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{Mempool: "mempool"}
}

type ModuleParams struct {
	ClientID        tt.ClientID
	Tps             int
	BufSize         int
	WatermarkWindow tt.TxNo
}

func DefaultModuleParams(clientID tt.ClientID) *ModuleParams {
	return &ModuleParams{
		ClientID:        clientID,
		Tps:             1,
		BufSize:         0,
		WatermarkWindow: 1000000,
	}
}

type LocalTXGen struct {
	modules *ModuleConfig
	params  *ModuleParams

	nextTXNo  tt.TxNo
	delivered *clientprogress.DeliveredTXs
	wmCond    *sync.Cond
	eventsOut chan *events.EventList
	stopChan  chan struct{}
	doneChan  chan struct{}
}

func New(moduleConfig *ModuleConfig, params *ModuleParams) *LocalTXGen {
	return &LocalTXGen{
		modules:   moduleConfig,
		params:    params,
		nextTXNo:  0,
		delivered: clientprogress.EmptyDeliveredTXs(logging.ConsoleInfoLogger),
		wmCond:    sync.NewCond(&sync.Mutex{}),
		eventsOut: make(chan *events.EventList, params.BufSize),
		stopChan:  make(chan struct{}),
		doneChan:  make(chan struct{}),
	}
}

func (gen *LocalTXGen) ImplementsModule() {}

// ApplyEvents returns an error on any event it receives, except fot the Init event, which it silently ignores.
func (gen *LocalTXGen) ApplyEvents(_ context.Context, evts *events.EventList) error {
	for _, evt := range evts.Slice() {
		if _, ok := evt.Type.(*eventpb.Event_Init); !ok {
			return fmt.Errorf("local request generator cannot apply events other than Init")
		}
	}
	return nil
}

// EventsOut returns the channel to which LocalTXGen writes all output events (in this case just NewRequests events).
func (gen *LocalTXGen) EventsOut() <-chan *events.EventList {
	return gen.eventsOut
}

// Start starts generating transactions at the defined rate.
func (gen *LocalTXGen) Start() {

	var cnt atomic.Int64

	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(gen.params.Tps))
		defer func() {
			ticker.Stop()
			close(gen.doneChan)
		}()

		for {
			select {
			case <-ticker.C:
				// Wait until the new transaction is within the client watermark window.
				gen.wmCond.L.Lock()
				// TODO: Dirty hack with .Pb(). Make LowWm accessible directly from DeliveredTXs.
				for tt.TxNo(gen.delivered.Pb().LowWm)+gen.params.WatermarkWindow < gen.nextTXNo {
					select {
					case <-gen.stopChan:
						gen.wmCond.L.Unlock()
						return
					default:
						gen.wmCond.Wait()
					}
				}
				gen.wmCond.L.Unlock()

				// Create new transaction event.
				tx := trantorpbtypes.Transaction{
					ClientId: gen.params.ClientID,
					TxNo:     gen.nextTXNo,
					Type:     0,
					Data:     []byte{0},
				}
				evts := events.ListOf(
					mempoolpbevents.NewTransactions(gen.modules.Mempool, []*trantorpbtypes.Transaction{&tx}).Pb(),
				)
				gen.nextTXNo++

				select {
				case gen.eventsOut <- evts:
					cnt.Inc()
				case <-gen.stopChan:
					return
				}
			case <-gen.stopChan:
				return
			}
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Printf("Submitted: %d\n", cnt.Swap(0))
		}
	}()
}

// Stop stops the transaction generation and waits until the internal transaction generation thread stops.
// A call to Stop without a prior call to Start blocks until Start is called (immediately receiving the stop signal).
// Note that, in such a case, the calls to Start and Stop will inevitably be concurrent,
// and some transactions might be submitted before Stop returns.
// After Stop returns, however, no transactions will be output.
func (gen *LocalTXGen) Stop() {
	gen.wmCond.L.Lock()
	close(gen.stopChan)
	gen.wmCond.Broadcast()
	gen.wmCond.L.Unlock()
	<-gen.doneChan
}

// ====================================================================================================
// Trantor static app interface
// ====================================================================================================

func (gen *LocalTXGen) ApplyTXs(txs []*trantorpbtypes.Transaction) error {
	gen.wmCond.L.Lock()
	for _, tx := range txs {
		gen.delivered.Add(tx.TxNo)
	}
	gen.delivered.GarbageCollect()
	gen.wmCond.Broadcast()
	gen.wmCond.L.Unlock()
	return nil
}

func (gen *LocalTXGen) Snapshot() ([]byte, error) {
	return []byte{0}, nil
}

func (gen *LocalTXGen) RestoreState(_ *checkpoint.StableCheckpoint) error {
	return nil
}

func (gen *LocalTXGen) Checkpoint(_ *checkpoint.StableCheckpoint) error {
	return nil
}
