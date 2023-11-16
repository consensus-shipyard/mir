package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/events"
	communicationpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/communicationpb/events"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb"
	tpmpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/tpmpb/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/go-errors/errors"
	"github.com/mitchellh/hashstructure"
)

const (
	EXPONENTIAL_MINUTE_FACTOR = 0.3
)

type blockRequest struct {
	HeadId  uint64
	Payload *blockchainpb.Payload
}

type minerModule struct {
	blockRequests chan blockRequest
	eventsOut     chan *events.EventList
	logger        logging.Logger
}

func NewMiner(logger logging.Logger) modules.ActiveModule {
	return &minerModule{
		blockRequests: make(chan blockRequest),
		eventsOut:     make(chan *events.EventList),
		logger:        logger,
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
		case *eventpb.Event_Miner:
			switch e := e.Miner.Type.(type) {
			case *minerpb.Event_BlockRequest:
				m.blockRequests <- blockRequest{e.BlockRequest.GetHeadId(), e.BlockRequest.GetPayload()}
				return nil
			case *minerpb.Event_NewHead:
				m.eventsOut <- events.ListOf(tpmpbevents.NewBlockRequest("tpm", e.NewHead.GetHeadId()).Pb())
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
		blockRequest := <-m.blockRequests
		cancel()                                               // abort ongoing mining
		ctx, cancel = context.WithCancel(context.Background()) // new context for new mining
		m.logger.Log(logging.LevelInfo, "New mining operation", "headId", formatBlockId(blockRequest.HeadId))
		go func() {
			delay := time.Duration(rand.ExpFloat64() * float64(time.Minute) * EXPONENTIAL_MINUTE_FACTOR)
			select {
			case <-ctx.Done():
				m.logger.Log(logging.LevelDebug, "Mining aborted", "headId", formatBlockId(blockRequest.HeadId))
				return
			case <-time.After(delay):
				block := &blockchainpb.Block{BlockId: 0, PreviousBlockId: blockRequest.HeadId, Payload: blockRequest.Payload, Timestamp: time.Now().Unix()}
				hash, err := hashstructure.Hash(block, nil)
				if err != nil {
					m.logger.Log(logging.LevelError, "Failed to hash block", "error", err)
					panic(err)
				}
				block.BlockId = hash
				// Block mined! Broadcast to blockchain and send event to bcm.
				m.logger.Log(logging.LevelInfo, "Block mined", "headId", formatBlockId(hash), "parentId", formatBlockId(blockRequest.HeadId))
				m.eventsOut <- events.ListOf(bcmpbevents.NewBlock("bcm", block).Pb(), communicationpbevents.NewBlock("communication", block).Pb())
				return
			}
		}()
	}
}
