package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb"
	applicationpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/events"
	bcmpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/events"
	broadcastpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/broadcastpb/events"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb"
	payloadpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb/types"
	blockchainpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/samples/blockchain/utils"
	"github.com/go-errors/errors"
	"github.com/mitchellh/hashstructure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/**
 * Miner module
 * ============
 *
 * The miner module is responsible for mining new blocks.
 * It simulates the process of mining a block by waiting for a random amount of time and then broadcasting the mined block.
 * This random amount of time is sampled from an exponential distribution with a mean of `expMinuteFactor` minutes.
 * The mining is orchestrated by a separate goroutine (mineWorkerManager), so that the miner module can continue to receive and process events.
 *
 * The operation of the miner module at a high level is as follows:
 * 1. When it is notified of a new head (NewHead event), it prepares to mine the next block by sending a PayloadRequest event to the application module.
 *	  If it is already mining a block, it aborts the ongoing mining operation.
 * 2. When it receives the PayloadResponse containing a payload for the next block, it starts mining a new block with the received payload.
 * 3. When it mines a new block, it broadcasts it to all other modules by sending a NewBlock message to the broadcast module.
 *    It also shares the block with the blockchain manager module (BCM) by sending a NewBlock event to it.
 */

/**
 * Request for a new block to be mined.
 */
type blockRequest struct {
	HeadId  uint64                  // id of the head of the blockchain (parent of the new block)
	Payload *payloadpbtypes.Payload // payload for the new block
}

type minerModule struct {
	nodeID          t.NodeID               // id of this node
	expMinuteFactor float64                // factor for exponential distribution for random mining duration
	blockRequests   chan blockRequest      // channel to send block requests to mineWorker
	eventsOut       chan *events.EventList // channel to send events to other modules
	logger          logging.Logger
}

func NewMiner(nodeID t.NodeID, expMinuteFactor float64, logger logging.Logger) modules.ActiveModule {
	return &minerModule{
		nodeID:          nodeID,
		expMinuteFactor: expMinuteFactor,
		blockRequests:   make(chan blockRequest),
		eventsOut:       make(chan *events.EventList),
		logger:          logger,
	}
}

func (m *minerModule) ImplementsModule() {}

func (m *minerModule) EventsOut() <-chan *events.EventList {
	return m.eventsOut
}

func (m *minerModule) ApplyEvents(context context.Context, eventList *events.EventList) error {
	for _, event := range eventList.Slice() {
		switch e := event.Type.(type) {
		case *eventpb.Event_Init:
			go m.mineWorkerManager()
		case *eventpb.Event_Application:
			switch e := e.Application.Type.(type) {
			case *applicationpb.Event_PayloadResponse:
				m.blockRequests <- blockRequest{e.PayloadResponse.HeadId, payloadpbtypes.PayloadFromPb(e.PayloadResponse.Payload)}
				return nil
			default:
				return errors.Errorf("unknown event: %T", e)
			}
		case *eventpb.Event_Miner:
			switch e := e.Miner.Type.(type) {
			case *minerpb.Event_NewHead:
				m.eventsOut <- events.ListOf(applicationpbevents.PayloadRequest("application", e.NewHead.HeadId).Pb())
			default:
				return errors.Errorf("unknown miner event: %T", e)
			}
		default:
			return errors.Errorf("unknown event: %T", e)
		}
	}

	return nil
}

func (m *minerModule) mineWorkerManager() {
	ctx, cancel := context.WithCancel(context.Background())

	for {
		blockRequest := <-m.blockRequests                      // blocking until a new block request is received
		cancel()                                               // abort ongoing mining
		ctx, cancel = context.WithCancel(context.Background()) // new context for new mining
		m.logger.Log(logging.LevelInfo, "New mining operation", "headId", utils.FormatBlockId(blockRequest.HeadId))

		// spawn new goroutine to mine the requested block
		go func() {
			delay := time.Duration(rand.ExpFloat64() * float64(time.Minute) * m.expMinuteFactor)
			select {
			case <-ctx.Done():
				// Mining aborted. Do nothing.
				m.logger.Log(logging.LevelDebug, "Mining aborted", "headId", utils.FormatBlockId(blockRequest.HeadId))
				return
			case <-time.After(delay):
				// Mining completed. Create block and broadcast.
				block := &blockchainpbtypes.Block{BlockId: 0, PreviousBlockId: blockRequest.HeadId, Payload: blockRequest.Payload, Timestamp: timestamppb.Now(), MinerId: m.nodeID}
				hash, err := hashstructure.Hash(block, nil)
				if err != nil {
					m.logger.Log(logging.LevelError, "Failed to hash block", "error", err)
					panic(err)
				}
				block.BlockId = hash
				// Send block to BCM and broadcast module
				m.logger.Log(logging.LevelInfo, "Block mined", "blockId", utils.FormatBlockId(hash), "parentId", utils.FormatBlockId(blockRequest.HeadId))
				m.eventsOut <- events.ListOf(bcmpbevents.NewBlock("bcm", block).Pb(), broadcastpbevents.NewBlock("broadcast", block).Pb())
				return
			}
		}()
	}
}
