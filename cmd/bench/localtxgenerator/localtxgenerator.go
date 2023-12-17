package localtxgenerator

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/cmd/bench/stats"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

type ModuleConfig struct {
	Mempool stdtypes.ModuleID
}

func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{Mempool: "mempool"}
}

type ModuleParams struct {
	ClientID    tt.ClientID
	PayloadSize int `json:",string"`
	NumClients  int `json:",string"`
}

func DefaultModuleParams(clientID tt.ClientID) ModuleParams {
	return ModuleParams{
		ClientID:    clientID,
		PayloadSize: 512,
		NumClients:  1,
	}
}

type LocalTXGen struct {
	modules       ModuleConfig
	params        ModuleParams
	statsTrackers []stats.Tracker
	txChan        chan stdtypes.Event
	clients       map[tt.ClientID]*client

	eventsOut chan *stdtypes.EventList
	wg        sync.WaitGroup
	stopChan  chan struct{}
}

func New(moduleConfig ModuleConfig, params ModuleParams) *LocalTXGen {
	txChan := make(chan stdtypes.Event, params.NumClients)
	logger := logging.ConsoleInfoLogger

	clients := make(map[tt.ClientID]*client, params.NumClients)
	for i := 0; i < params.NumClients; i++ {
		clID := tt.ClientID(fmt.Sprintf("%v.%d", params.ClientID, i))
		clients[clID] = newClient(clID, moduleConfig, params, txChan, logger)
	}

	return &LocalTXGen{
		modules:   moduleConfig,
		params:    params,
		txChan:    txChan,
		clients:   clients,
		eventsOut: make(chan *stdtypes.EventList),
		stopChan:  make(chan struct{}),
	}
}

func (gen *LocalTXGen) TrackStats(tracker stats.Tracker) {
	gen.statsTrackers = append(gen.statsTrackers, tracker)
	for _, c := range gen.clients {
		c.TrackStats(tracker)
	}
}

// Start starts generating transactions at the defined rate.
func (gen *LocalTXGen) Start() {

	var cnt atomic.Int64

	for _, c := range gen.clients {
		c.Start(&cnt)
	}

	gen.wg.Add(2)

	// Shovel transactions from individual clients to the module output
	go func() {
		defer gen.wg.Done()

		txEventList := stdtypes.EmptyList()

		for {
			// Give priority to collecting submitted transactions.
			select {
			case txEvent := <-gen.txChan:
				txEventList.PushBack(txEvent)
			default:
				// Only if there is no new transaction to be submitted and added to the list,
				// try to submit the whole event list to Trantor.
				select {
				case txEvent := <-gen.txChan:
					txEventList.PushBack(txEvent)
				case gen.eventsOut <- txEventList:
					txEventList = stdtypes.EmptyList()
				case <-gen.stopChan:
					return
				}
			}
		}
	}()

	// Periodically print the number of submitted transactions
	go func() {
		defer gen.wg.Done()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf("Submitted: %d\n", cnt.Swap(0))
			case <-gen.stopChan:
				return
			}
		}
	}()

}

// Stop stops the transaction generation and waits until the internal transaction generation thread stops.
// A call to Stop without a prior call to Start blocks until Start is called (immediately receiving the stop signal).
// Note that, in such a case, the calls to Start and Stop will inevitably be concurrent,
// and some transactions might be submitted before Stop returns.
// After Stop returns, however, no transactions will be output.
func (gen *LocalTXGen) Stop() {
	close(gen.stopChan)
	for _, c := range gen.clients {
		c.Stop()
	}
	gen.wg.Wait()
}

// ====================================================================================================
// Active module interface
// ====================================================================================================

func (gen *LocalTXGen) ImplementsModule() {}

// ApplyEvents returns an error on any event it receives, except fot the Init event, which it silently ignores.
func (gen *LocalTXGen) ApplyEvents(_ context.Context, evts *stdtypes.EventList) error {
	for _, evt := range evts.Slice() {

		// We only support proto events.
		pbevent, ok := evt.(*eventpb.Event)
		if !ok {
			return es.Errorf("Timer only supports proto events, but received %T", evt)
		}

		// Complain about anything else than an Init event.
		if _, ok := pbevent.Type.(*eventpb.Event_Init); !ok {
			return fmt.Errorf("local request generator cannot apply events other than Init")
		}
	}
	return nil
}

// EventsOut returns the channel to which LocalTXGen writes all output events (in this case just NewRequests events).
func (gen *LocalTXGen) EventsOut() <-chan *stdtypes.EventList {
	return gen.eventsOut
}

// ====================================================================================================
// Trantor static app interface
// ====================================================================================================

func (gen *LocalTXGen) ApplyTXs(txs []*trantorpbtypes.Transaction) error {
	for _, tx := range txs {
		if c, ok := gen.clients[tx.ClientId]; ok {
			c.Deliver(tx)
		}
	}
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
