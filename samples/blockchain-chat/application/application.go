package application

// app module

import (
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/blockchain/utils"
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
	"github.com/filecoin-project/mir/samples/blockchain-chat/application/payloads"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/**
* Application module
* ==================
*
* The application module is reponsible for performing the actual application logic and to interact with users or other applications.
* It does not hold any state, but instead relies on the blockchain manager module (BCM) to store the state.
* However, the application is responsible for computing the state given a chain of blocks and a state associated with the first block in the chain.
* Also, the application module manages payloads and must provide payloads for new blocks to the miner.
*
* The application module must perform the following tasks:
* 1. Initialize the blockchain by sending the initial state to the BCM in an InitBlockchain event.
* 2. When it receives a PayloadRequest event, it must provide a payload for the next block.
*	 Even if no payloads are available, a payload **must** be provided; however, this payload can be empty.
* 3. When it receives a HeadChange event, it must compute the state at the new head of the blockchain.
*    This state is then registered with the BCM by sending it a RegisterCheckpoint event.
*	 (A checkpoint is a block stored by the BCM that has a state stored with it.)
*    Additionally, the information provided in the _HeadChange_ event might be useful for the payload management.
* 4. When it receives a VerifyBlocksRequest event, it must verify that the given chain is valid at an application level and respond with a VerifyBlocksResponse event.
*    Whether or not the blocks link together correctly is verified by the BCM.
*
* This application module implements a simple chat application.
* It takes new messages from the user (MessageInput event) and combines them with a sender id and "sent" timestamp as payloads.
* These payloads are stored in the payload manager (see applicaion/payloads/payloads.go).
* The state is the list of all messages that have been sent and timestamps for when each sender last sent a message.
* At the application level, a chain is valid if the timestamps are monotonically increasing for each sender.
 */

type ApplicationModule struct {
	m      *dsl.Module
	pm     *payloads.PayloadManager // store for payloads
	nodeID t.NodeID                 // to set sender in payload
	logger logging.Logger
}

// application-application events

/**
* Helper functions
 */

// helper function that applies a block's payload to the given state, thereby computing the new state
func applyBlockToState(state *statepbtypes.State, block *blockchainpbtypes.Block) *statepbtypes.State {
	sender := block.Payload.Sender
	timeStamps := state.LastSentTimestamps
	msgHistory := state.MessageHistory

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

/**
* Mir event handlers
 */

// TODO: rename - fork is confusing

// Handler called by BCM when the head changes.
// It handles cases where the head changes because the canonical chain was extended as well as cases where the canonical chain changes because of a fork.
func (am *ApplicationModule) handleHeadChange(removedChain, addedChain, checkpointToForkRootChain []*blockchainpbtypes.Block, checkpointState *statepbtypes.State) error {
	am.logger.Log(logging.LevelInfo, "Processing fork update", "poolSize", am.pm.PoolSize())

	// add "remove chain" transactions to pool
	for _, block := range removedChain {
		am.pm.AddPayload(block.Payload)
	}

	// remove "add chain" transactions from pool
	for _, block := range addedChain {
		am.pm.RemovePayload(block.Payload)
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

	am.logger.Log(logging.LevelInfo, "Pool after fork", "poolSize", am.pm.PoolSize())

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

// Handler called by BCM when a chain needs to be verified.
// The checks are purely at the application level.
// In this implementation, it simply checks that the timestamps are monotonically increasing for each sender.
func (am *ApplicationModule) handleVerifyBlocksRequest(checkpointState *statepbtypes.State, chainCheckpointToStart, chainToVerify []*blockchainpbtypes.Block) error {
	am.logger.Log(logging.LevelDebug, "Processing verify blocks request", "checkpoint", utils.FormatBlockId(chainCheckpointToStart[0].BlockId), "first block to verify", utils.FormatBlockId(chainToVerify[0].BlockId), "last block to verify", utils.FormatBlockId(chainToVerify[len(chainToVerify)-1].BlockId))

	timestampMap := make(map[t.NodeID]time.Time)
	for _, lt := range checkpointState.LastSentTimestamps {
		timestampMap[lt.NodeId] = lt.Timestamp.AsTime()
	}

	chain := append(chainCheckpointToStart, chainToVerify...) // chainCheckpointToStart wouldn't need to be verified, but we need it to get the timestamps

	for _, block := range chain {
		blockTs := block.Payload.Timestamp.AsTime()
		// verify block
		if ts, ok := timestampMap[block.Payload.Sender]; ok {
			if ts.After(blockTs) {
				// block in chain invalid, don't respond
				return nil
			}
		}

		// update timestamp
		timestampMap[block.Payload.Sender] = blockTs
	}

	// if no blocks are invalid, responds with chain
	applicationpbdsl.VerifyBlocksResponse(*am.m, "bcm", chainToVerify)

	return nil
}

// Handler called by the miner when it needs a payload for the next block.
func (am *ApplicationModule) handlePayloadRequest(head_id uint64) error {
	am.logger.Log(logging.LevelDebug, "Processing payload request", "headId", utils.FormatBlockId(head_id))

	payload := am.pm.GetPayload()

	// if no payload is available, create empty payload
	// this is important to ensure that all nodes are always mining new blocks
	if payload == nil {
		payload = &payloadpbtypes.Payload{
			Message:   "",
			Timestamp: timestamppb.Now(),
			Sender:    am.nodeID,
		}
	}

	applicationpbdsl.PayloadResponse(*am.m, "miner", head_id, payload)

	return nil
}

func NewApplication(nodeID t.NodeID, logger logging.Logger) modules.Module {

	m := dsl.NewModule("application")
	am := &ApplicationModule{
		m:      &m,
		logger: logger,
		nodeID: nodeID,
		pm:     payloads.New(),
	}

	dsl.UponInit(m, func() error {
		// init blockchain
		bcmpbdsl.InitBlockchain(*am.m, "bcm", &statepbtypes.State{
			MessageHistory:     []string{},
			LastSentTimestamps: []*statepbtypes.LastSentTimestamp{},
		})

		return nil
	})

	applicationpbdsl.UponPayloadRequest(m, am.handlePayloadRequest)
	applicationpbdsl.UponHeadChange(m, am.handleHeadChange)
	applicationpbdsl.UponVerifyBlocksRequest(m, am.handleVerifyBlocksRequest)

	applicationpbdsl.UponMessageInput(m, func(text string) error {
		am.pm.AddPayload(&payloadpbtypes.Payload{
			Message:   text,
			Timestamp: timestamppb.Now(),
			Sender:    am.nodeID,
		})

		return nil
	})

	return m
}
