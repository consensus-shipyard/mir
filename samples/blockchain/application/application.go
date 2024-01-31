package application

// app module

import (
	"fmt"

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
	"google.golang.org/protobuf/types/known/timestamppb"
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

	// why is this necessary? it should be verified already
	for i, lt := range timeStamps {
		if lt.NodeId == sender {
			ltTimestamp := lt.Timestamp.AsTime()
			if ltTimestamp.After(block.Payload.Timestamp.AsTime()) {
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

func (am *ApplicationModule) handleForkUpdate(removedChain, addedChain, checkpointToForkRootChain []*blockchainpbtypes.Block, checkpointState *statepbtypes.State) error {
	am.logger.Log(logging.LevelInfo, "Processing fork update", "poolSize", am.tm.PoolSize())

	// add "remove chain" transactions to pool
	for _, block := range removedChain {
		am.tm.AddPayload(block.Payload)
	}

	// remove "add chain" transactions from pool
	for _, block := range addedChain {
		am.tm.RemovePayload(block.Payload)
	}

	state := checkpointState
	// compute state at fork roo
	// skip first as checkpoint state already 'included' its payload
	for _, block := range checkpointToForkRootChain[1:] {
		state = applyBlockToState(state, block)
	}
	// compute state at new head
	for _, block := range addedChain {
		state = applyBlockToState(state, block)
	}

	am.logger.Log(logging.LevelInfo, "Pool after fork", "poolSize", am.tm.PoolSize())

	// register checkpoint
	blockId := addedChain[len(addedChain)-1].BlockId
	bcmpbdsl.RegisterCheckpoint(*am.m, "bcm", blockId, state)
	interceptorpbdsl.StateUpdate(*am.m, "devnull", state)

	// print state
	fmt.Printf("=== STATE ===\n")
	for _, msg := range state.MessageHistory {
		fmt.Println(msg)
	}
	fmt.Printf("=============\nEnter new message: \n")

	return nil
}

func (am *ApplicationModule) handleVerifyBlocksRequest(checkpointState *statepbtypes.State, chainCheckpointToStart, chainToVerify []*blockchainpbtypes.Block) error {
	am.logger.Log(logging.LevelDebug, "Processing verify block request")

	timeStamps := checkpointState.LastSentTimestamps

	chain := append(chainCheckpointToStart, chainToVerify...) // chainCheckpointToStart wouldn't need to be verified, but we need it to get the timestamps

	for _, block := range chain {
		blockTs := block.Payload.Timestamp.AsTime()
		// verify block
		for i, lt := range timeStamps {
			ltTimestamp := lt.Timestamp.AsTime()
			if lt.NodeId == block.Payload.Sender {
				if ltTimestamp.After(blockTs) {
					// block in chain invalid, don't respond
					return nil
				}

				// remove old timestamp, if it exists
				timeStamps[i] = timeStamps[len(timeStamps)-1]
				timeStamps = timeStamps[:len(timeStamps)-1]
			}

			// add new timestamp
			timeStamps = append(timeStamps, &statepbtypes.LastSentTimestamp{
				NodeId:    block.Payload.Sender,
				Timestamp: block.Payload.Timestamp,
			})
		}
	}

	// if no blocks are invalid, responds with chain
	applicationpbdsl.VerifyBlocksResponse(*am.m, "bcm", chainToVerify)

	return nil
}

// transaction management events

func (am *ApplicationModule) handlePayloadRequest(head_id uint64) error {
	am.logger.Log(logging.LevelDebug, "Processing payload request", "headId", utils.FormatBlockId(head_id))

	payload := am.tm.GetPayload()

	if payload == nil {
		payload = &payloadpbtypes.Payload{
			Message:   "",
			Timestamp: timestamppb.Now(),
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
	applicationpbdsl.UponVerifyBlocksRequest(m, am.handleVerifyBlocksRequest)

	applicationpbdsl.UponMessageInput(m, func(text string) error {
		am.tm.AddPayload(&payloadpbtypes.Payload{
			Message:   text,
			Timestamp: timestamppb.Now(),
			Sender:    am.nodeID,
		})

		return nil
	})

	return m
}
