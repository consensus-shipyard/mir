package localtxgenerator

import (
	"encoding/binary"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/cmd/bench/stats"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	mempoolpbevents "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

type client struct {
	id            tt.ClientID
	modules       ModuleConfig
	params        ModuleParams
	randSource    *rand.Rand
	nextTXNo      tt.TxNo
	statsTrackers []stats.Tracker
	txOutChan     chan *eventpb.Event
	txDeliverChan chan *trantorpbtypes.Transaction
	stopChan      chan struct{}
	wg            sync.WaitGroup
	logger        logging.Logger
}

func newClient(id tt.ClientID, moduleConfig ModuleConfig, params ModuleParams, txOutChan chan *eventpb.Event, logger logging.Logger) *client {
	seed := make([]byte, 8)
	copy(seed, id.Bytes())
	return &client{
		id:            id,
		modules:       moduleConfig,
		params:        params,
		randSource:    rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(seed)))), // nolint:gosec
		nextTXNo:      0,
		statsTrackers: make([]stats.Tracker, 0),
		txOutChan:     txOutChan,
		txDeliverChan: make(chan *trantorpbtypes.Transaction, 1),
		stopChan:      make(chan struct{}),
		logger:        logger,
	}
}

// Start starts generating transactions at the defined rate.
func (c *client) Start(cnt *atomic.Int64) {

	c.wg.Add(1)
	go func() {
		c.wg.Done()

		for {
			tx := c.newTX()
			evt := mempoolpbevents.NewTransactions(c.modules.Mempool, []*trantorpbtypes.Transaction{tx}).Pb()

			// Track the submission of the new transaction for statistics.
			for _, statsTracker := range c.statsTrackers {
				statsTracker.Submit(tx)
			}

			select {
			case c.txOutChan <- evt:
				cnt.Add(1)
			case <-c.stopChan:
				return
			}

			select {
			case deliveredTx := <-c.txDeliverChan:
				if err := c.registerDelivery(tx, deliveredTx); err != nil {
					c.logger.Log(logging.LevelError, "Error registering transaction delivery.", "err", err)
				}
			case <-c.stopChan:
				return
			}
		}
	}()
}

func (c *client) Stop() {
	close(c.stopChan)
	c.wg.Wait()
}

func (c *client) Deliver(tx *trantorpbtypes.Transaction) {
	c.txDeliverChan <- tx
}

func (c *client) TrackStats(tracker stats.Tracker) {
	c.statsTrackers = append(c.statsTrackers, tracker)
}

func (c *client) registerDelivery(submitted, delivered *trantorpbtypes.Transaction) error {

	// Sanity check that the submitted transaction was delivered.
	if !reflect.DeepEqual(submitted, delivered) {
		return es.Errorf("delivered transaction is not the submitted one: %v, %v", submitted, delivered)
	}

	// Track delivery statistics.
	for _, statsTracker := range c.statsTrackers {
		statsTracker.Deliver(delivered)
	}
	return nil
}

func (c *client) newTX() *trantorpbtypes.Transaction {
	defer func() { c.nextTXNo++ }()

	// Generate random transaction payload.
	data := make([]byte, c.params.PayloadSize)
	c.randSource.Read(data)

	// Create new transaction event.
	return &trantorpbtypes.Transaction{
		ClientId: c.id,
		TxNo:     c.nextTXNo,
		Type:     0,
		Data:     data,
	}
}
