package application

// app module

import (
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	applicationpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/dsl"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	interceptorpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/interceptorpb/dsl"
	payloadpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb/types"
	statepbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb/types"
	blockchainpbtypes "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/samples/blockchain/application/config"
	"github.com/filecoin-project/mir/samples/blockchain/application/transactions"
	"github.com/filecoin-project/mir/samples/blockchain/utils"
)

type ApplicationModule struct {
	m      *dsl.Module
	logger logging.Logger
	tm     *transactions.TransactionManager
	nodeID t.NodeID
}

// application-application events

func applyBlockToState(state *statepbtypes.State, block *blockchainpbtypes.Block) *statepbtypes.State {
	sender := block.Payload.Sender
	timeStamps := state.LastSentTimestamps
	msgHistory := state.MessageHistory

	for i, lt := range timeStamps {
		if lt.NodeId == sender {
			if lt.Timestamp > block.Payload.Timestamp {
				panic("invalid ordering - there is a block that should never have been accepted")
			}

			// remove old timestamp, if it exists
			timeStamps[i] = timeStamps[len(state.LastSentTimestamps)-1]
			timeStamps = timeStamps[:len(state.LastSentTimestamps)-1]
		}
	}

	timeStamps = append(timeStamps, &statepbtypes.LastSentTimestamp{
		NodeId:    sender,
		Timestamp: block.Payload.Timestamp,
	})

	// empty block, just skip
	if block.Payload.Message != "" {
		msgHistory = append(msgHistory, block.Payload.Message)
	}

	return &statepbtypes.State{
		MessageHistory:     msgHistory,
		LastSentTimestamps: timeStamps,
	}
}

func (am *ApplicationModule) handleForkUpdate(removedChain, addedChain *blockchainpbtypes.Blockchain, forkState *statepbtypes.State) error {
	am.logger.Log(logging.LevelInfo, "Processing fork update", "poolSize", am.tm.PoolSize())

	// add "remove chain" transactions to pool
	for _, block := range removedChain.Blocks {
		am.tm.AddPayload(block.Payload)
	}

	// remove "add chain" transactions from pool
	for _, block := range addedChain.Blocks {
		am.tm.RemovePayload(block.Payload)
	}

	// apply state to fork state
	state := forkState
	for _, block := range addedChain.Blocks {
		state = applyBlockToState(state, block)
	}

	am.logger.Log(logging.LevelInfo, "Pool after fork", "poolSize", am.tm.PoolSize())

	// register checkpoint
	blockId := addedChain.Blocks[len(addedChain.Blocks)-1].BlockId
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

	if payload == nil {
		payload = &payloadpbtypes.Payload{
			Message:   "",
			Timestamp: time.Now().Unix(),
			Sender:    am.nodeID,
		}
	}

	applicationpbdsl.PayloadResponse(*am.m, "miner", head_id, payload) // not using head id anywhere so we can get rid of it

	return nil
}

func NewApplication(logger logging.Logger, nodeID t.NodeID) modules.PassiveModule {

	m := dsl.NewModule("application")
	am := &ApplicationModule{
		m:      &m,
		logger: logger,
		nodeID: nodeID,
		tm:     transactions.New(),
	}

	dsl.UponInit(m, func() error {
		// init blockchain
		bcmpbdsl.InitBlockchain(*am.m, "bcm", config.InitialState)

		return nil
	})

	applicationpbdsl.UponPayloadRequest(m, am.handlePayloadRequest)
	applicationpbdsl.UponForkUpdate(m, am.handleForkUpdate)

	applicationpbdsl.UponMessageInput(m, func(text string) error {
		am.tm.AddPayload(&payloadpbtypes.Payload{
			Message:   text,
			Timestamp: time.Now().Unix(),
			Sender:    am.nodeID,
		})

		return nil
	})

	return m
}
