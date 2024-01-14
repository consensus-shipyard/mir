package application

// app module

import (
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	applicationpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/dsl"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	interceptorpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/interceptorpb/dsl"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb"
	"github.com/filecoin-project/mir/samples/blockchain/application/config"
	"github.com/filecoin-project/mir/samples/blockchain/application/transactions"
	"github.com/filecoin-project/mir/samples/blockchain/utils"
)

type ApplicationModule struct {
	m      *dsl.Module
	logger logging.Logger
	tm     *transactions.TransactionManager
	name   string
}

// application-application events

func applyBlockToState(state *statepb.State, block *blockchainpb.Block) *statepb.State {
	// empty block, just skip
	if block.Payload.Message == "" {
		return state
	}
	return &statepb.State{
		MessageHistory: append(state.MessageHistory, block.Payload.Message),
	}
}

func (am *ApplicationModule) handleForkUpdate(removedChain, addedChain *blockchainpb.Blockchain, forkState *statepb.State) error {
	am.logger.Log(logging.LevelInfo, "Processing fork update", "poolSize", am.tm.PoolSize())

	// add "remove chain" transactions to pool
	for _, block := range removedChain.GetBlocks() {
		am.tm.AddPayload(block.Payload)
	}

	// remove "add chain" transactions from pool
	for _, block := range addedChain.GetBlocks() {
		am.tm.RemovePayload(block.Payload)
	}

	// apply state to fork state
	state := forkState
	for _, block := range addedChain.GetBlocks() {
		state = applyBlockToState(state, block)
	}

	am.logger.Log(logging.LevelInfo, "Pool after fork", "poolSize", am.tm.PoolSize())

	// register checkpoint
	blockId := addedChain.GetBlocks()[len(addedChain.GetBlocks())-1].BlockId
	bcmpbdsl.RegisterCheckpoint(*am.m, "bcm", blockId, state)
	interceptorpbdsl.AppUpdate(*am.m, "devnull", state)

	// print state
	fmt.Printf("=== STATE ===\n")
	for _, msg := range state.MessageHistory {
		fmt.Println(msg)
	}
	fmt.Printf("=============\nEnter new message: \n")

	return nil
}

// transaction management events

func (am *ApplicationModule) handlePayloadRequest(head_id uint64) error {
	am.logger.Log(logging.LevelDebug, "Processing payload request", "headId", utils.FormatBlockId(head_id))

	payload := am.tm.GetPayload()

	applicationpbdsl.PayloadResponse(*am.m, "miner", head_id, payload) // not using head id anywhere so we can get rid of it

	return nil
}

func NewApplication(logger logging.Logger, name string) modules.PassiveModule {

	m := dsl.NewModule("application")
	am := &ApplicationModule{
		m:      &m,
		logger: logger,
		name:   name,
		tm:     transactions.New(name),
	}

	dsl.UponInit(m, func() error {
		// init blockchain
		bcmpbdsl.InitBlockchain(*am.m, "bcm", config.InitialState)

		return nil
	})

	applicationpbdsl.UponPayloadRequest(m, am.handlePayloadRequest)
	applicationpbdsl.UponForkUpdate(m, am.handleForkUpdate)

	applicationpbdsl.UponMessageInput(m, func(text string) error {
		am.tm.AddPayload(&payloadpb.Payload{
			Message:   text,
			Timestamp: time.Now().Unix(),
		})

		return nil
	})

	return m
}
