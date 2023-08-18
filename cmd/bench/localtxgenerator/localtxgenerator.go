package localtxgenerator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	es "github.com/go-errors/errors"
	"go.uber.org/atomic"

	"github.com/filecoin-project/mir/cmd/bench/stats"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	mempoolpbevents "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type ModuleConfig struct {
	Mempool t.ModuleID
}

func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{Mempool: "mempool"}
}

type ModuleParams struct {
	ClientID        tt.ClientID
	Tps             int     `json:",string"`
	PayloadSize     int     `json:",string"`
	BufSize         int     `json:",string"`
	WatermarkWindow tt.TxNo `json:",string"`
}

func DefaultModuleParams(clientID tt.ClientID) ModuleParams {
	return ModuleParams{
		ClientID:        clientID,
		Tps:             1,
		PayloadSize:     512,
		BufSize:         128,
		WatermarkWindow: 5000,
	}
}

type LocalTXGen struct {
	modules       ModuleConfig
	params        ModuleParams
	statsTrackers []stats.Tracker

	nextTXNo  tt.TxNo
	delivered *clientprogress.DeliveredTXs
	wmCond    *sync.Cond
	eventsOut chan *events.EventList
	stopChan  chan struct{}
	doneChan  chan struct{}
}

func New(moduleConfig ModuleConfig, params ModuleParams) *LocalTXGen {
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

func (gen *LocalTXGen) TrackStats(tracker stats.Tracker) {
	gen.statsTrackers = append(gen.statsTrackers, tracker)
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

				// Create a new transaction.
				tx := newTX(gen.params.ClientID, gen.nextTXNo, gen.params.PayloadSize)
				evts := events.ListOf(
					mempoolpbevents.NewTransactions(gen.modules.Mempool, []*trantorpbtypes.Transaction{tx}).Pb(),
				)

				// Track the submission of the new transaction for statistics.
				for _, statsTracker := range gen.statsTrackers {
					statsTracker.Submit(tx)
				}

				// Submit the transaction to Trantor.
				select {
				case gen.eventsOut <- evts:
					cnt.Inc()
				case <-gen.stopChan:
					return
				}

				// Increment the next transaction number.
				// ATTENTION: This has to happen after the transaction has been pushed to Mir,
				//   since we are using this value for checking if all transactions have been delivered.
				//   If we increment it before the transaction is pushed to Mir and the TX generation
				//   is stopped, the transaction will never be delivered and the Wait function will get stuck
				//   waiting for it forever.
				//   If we reach this line, however, we are sure to have submitted the transaction
				//   and Trantor guarantees its eventual delivery.
				gen.nextTXNo++

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

// Wait blocks until all submitted transactions have been delivered or ctx is canceled.
func (gen *LocalTXGen) Wait(ctx context.Context) error {
	gen.wmCond.L.Lock()
	defer gen.wmCond.L.Unlock()

	// Abort the waiting after context is done.
	abort := atomic.NewBool(false)
	go func() {
		<-ctx.Done()
		abort.Store(true)
		gen.wmCond.Broadcast()
	}()

	// Wait until all transactions have been delivered or waiting is aborted.
	// TODO: Dirty hack with .Pb(). Make LowWm accessible directly from DeliveredTXs.
	for tt.TxNo(gen.delivered.Pb().LowWm) < gen.nextTXNo && !abort.Load() {
		gen.wmCond.Wait()
	}

	// If the waiting was aborted, return an error.
	if abort.Load() {
		return es.Errorf(fmt.Sprintf("context canceled with nextTxNo=%d, lowWM=%d, and delivered=%v",
			gen.nextTXNo, gen.delivered.Pb().LowWm, gen.delivered.Pb().Delivered))
	}
	return nil
}

// ====================================================================================================
// Active module interface
// ====================================================================================================

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

// ====================================================================================================
// Trantor static app interface
// ====================================================================================================

func (gen *LocalTXGen) ApplyTXs(txs []*trantorpbtypes.Transaction) error {
	gen.wmCond.L.Lock()
	for _, tx := range txs {
		gen.delivered.Add(tx.TxNo)
		if tx.ClientId == gen.params.ClientID {
			for _, statsTracker := range gen.statsTrackers {
				statsTracker.Deliver(tx)
			}
		}
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

// ====================================================================================================
// Auxiliary functions
// ====================================================================================================

func newTX(clID tt.ClientID, txNo tt.TxNo, payloadSize int) *trantorpbtypes.Transaction {
	// Generate random transaction payload.
	data := make([]byte, payloadSize)
	rand.Read(data) //nolint:gosec

	// Create new transaction event.
	return &trantorpbtypes.Transaction{
		ClientId: clID,
		TxNo:     txNo,
		Type:     0,
		Data:     data,
	}
}
