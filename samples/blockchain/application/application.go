package application

// app module

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	applicationpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/dsl"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	interceptorpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/interceptorpb/dsl"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb"
	"github.com/filecoin-project/mir/samples/blockchain/application/transactions"
	"github.com/filecoin-project/mir/samples/blockchain/utils"
)

type localState struct {
	head  uint64
	state *statepb.State
}

type ApplicationModule struct {
	m            *dsl.Module
	currentState *localState
	logger       logging.Logger
	tm           *transactions.TransactionManager
	name         string
}

// application-application events

func applyBlockToState(state *statepb.State, block *blockchainpb.Block) *statepb.State {
	return &statepb.State{
		MessageHistory: append(state.MessageHistory, block.Payload.Message),
	}
}

// func (am *ApplicationModule) handleGetHeadToCheckpointChainResponse(requestID string, chain []*blockchainpb.BlockInternal) error {
// 	// TODO: ignoring request ids rn - they are probably not even needed
// 	am.logger.Log(logging.LevelInfo, "Received chain", "requestId", requestID, "chain", chain)

// 	if len(chain) == 0 {
// 		am.logger.Log(logging.LevelError, "Received empty chain - this should not happen")
// 		panic("Received empty chain - this should not happen")
// 	}

// 	state := chain[0].State
// 	blockId := chain[0].Block.BlockId
// 	if state == nil {
// 		am.logger.Log(logging.LevelError, "Received chain with empty checkpoint state - this should not happen")
// 		panic("Received chain with empty checkpoint state - this should not happen")
// 	}

// 	for _, block := range chain[1:] {
// 		state = applyBlockToState(state, block.Block)
// 		blockId = block.Block.BlockId
// 	}

// 	bcmpbdsl.RegisterCheckpoint(*am.m, "bcm", blockId, state)
// 	interceptorpbdsl.AppUpdate(*am.m, "devnull", state)
// 	am.currentState = &localState{
// 		head:  blockId,
// 		state: state,
// 	}

// 	return nil
// }

// func (am *ApplicationModule) handleNewHead(head_id uint64) error {
// 	am.logger.Log(logging.LevelDebug, "Processing new head, sending request for state update", "headId", utils.FormatBlockId(head_id))
// 	bcmpbdsl.GetHeadToCheckpointChainRequest(*am.m, "bcm", fmt.Sprint(head_id), "application")
// 	return nil
// }

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
	am.currentState = &localState{
		head:  blockId,
		state: state,
	}

	return nil
}

// transaction management events

func (am *ApplicationModule) handlePayloadRequest(head_id uint64) error {
	// NOTE: reject request for head that doesn't match current head?
	am.logger.Log(logging.LevelDebug, "Processing payload request", "headId", utils.FormatBlockId(head_id))

	payload := am.tm.GetPayload()
	applicationpbdsl.PayloadResponse(*am.m, "miner", head_id, payload)

	return nil
}

func NewApplication(logger logging.Logger, name string) modules.PassiveModule {

	m := dsl.NewModule("application")
	am := &ApplicationModule{m: &m, currentState: nil, logger: logger, name: name, tm: transactions.New(name)}

	dsl.UponInit(m, func() error {
		return nil
	})

	applicationpbdsl.UponPayloadRequest(m, am.handlePayloadRequest)
	applicationpbdsl.UponForkUpdate(m, am.handleForkUpdate)
	// applicationpbdsl.UponNewHead(m, am.handleNewHead)
	// bcmpbdsl.UponGetHeadToCheckpointChainResponse(m, am.handleGetHeadToCheckpointChainResponse)

	return m
}
