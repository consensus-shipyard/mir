package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/events"
	communicationpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/communicationpb/events"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/go-errors/errors"
	"github.com/mitchellh/hashstructure"
)

type blockRequest struct {
	HeadId  uint64
	Payload *blockchainpb.Payload
}

type minerModule struct {
	blockRequests chan blockRequest
	eventsOut     chan *events.EventList
}

func NewMiner() modules.ActiveModule {
	return &minerModule{
		blockRequests: make(chan blockRequest),
		eventsOut:     make(chan *events.EventList),
	}
}

func (m *minerModule) ImplementsModule() {}

func (m *minerModule) EventsOut() <-chan *events.EventList {
	return m.eventsOut
}

func (m *minerModule) ApplyEvents(context context.Context, events *events.EventList) error {
	for _, event := range events.Slice() {
		switch e := event.Type.(type) {
		case *eventpb.Event_Init:
			go m.mineWorkerManager()
		case *eventpb.Event_Miner:
			switch e := e.Miner.Type.(type) {
			case *minerpb.Event_BlockRequest:
				m.blockRequests <- blockRequest{e.BlockRequest.HeadId, e.BlockRequest.Payload}
				return nil
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
		println("Mining block on top of", blockRequest.HeadId)
		go func() {
			delay := time.Duration(rand.ExpFloat64() * float64(time.Minute))
			select {
			case <-ctx.Done():
				println("##### Mining aborted #####")
				return
			case <-time.After(delay):
				println("Block mined!")
				block := &blockchainpb.Block{BlockId: 0, PreviousBlockId: blockRequest.HeadId, Payload: blockRequest.Payload, Timestamp: time.Now().Unix()}
				hash, err := hashstructure.Hash(block, nil)
				if err != nil {
					panic(err)
				}
				block.BlockId = hash
				// Block mined! Broadcast to blockchain and send event to bcm.
				m.eventsOut <- events.ListOf(bcmpbevents.NewBlock("bcm", block).Pb(), communicationpbevents.NewBlock("communication", block).Pb())
				println("Block shared!")
				return
			}
		}()
	}
}
