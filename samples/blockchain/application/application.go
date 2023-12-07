package application

// app module

import (
	"fmt"
	"math/rand"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	applicationpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/dsl"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	interceptorpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/interceptorpb/dsl"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb"
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
}

func applyBlockToState(state *statepb.State, block *blockchainpb.Block) *statepb.State {
	return &statepb.State{
		Counter: state.Counter + block.Payload.AddMinus,
	}
}

func (am *ApplicationModule) handlePayloadRequest(head_id uint64) error {
	// NOTE: reject request for head that doesn't match current head?
	am.logger.Log(logging.LevelDebug, "Processing payload request", "headId", utils.FormatBlockId(head_id))

	var summand int64
	if rand.Intn(2) == 0 {
		summand = -1
	} else {
		summand = 1
	}

	applicationpbdsl.PayloadResponse(*am.m, "miner", head_id, &payloadpb.Payload{AddMinus: summand})

	return nil
}

func (am *ApplicationModule) handleGetHeadToCheckpointChainResponse(requestID string, chain []*blockchainpb.BlockInternal) error {
	// TODO: ignoring request ids rn - they are probably not even needed
	am.logger.Log(logging.LevelInfo, "Received chain", "requestId", requestID, "chain", chain)

	if len(chain) == 0 {
		am.logger.Log(logging.LevelError, "Received empty chain - this should not happen")
		panic("Received empty chain - this should not happen")
	}

	state := chain[0].State
	blockId := chain[0].Block.BlockId
	if state == nil {
		am.logger.Log(logging.LevelError, "Received chain with empty checkpoint state - this should not happen")
		panic("Received chain with empty checkpoint state - this should not happen")
	}

	for _, block := range chain[1:] {
		state = applyBlockToState(state, block.Block)
		blockId = block.Block.BlockId
	}

	bcmpbdsl.RegisterCheckpoint(*am.m, "bcm", blockId, state)
	interceptorpbdsl.AppUpdate(*am.m, "devnull", state.Counter)
	am.currentState = &localState{
		head:  blockId,
		state: state,
	}

	return nil
}

func (am *ApplicationModule) handleNewHead(head_id uint64) error {
	am.logger.Log(logging.LevelDebug, "Processing new head, sending request for state update", "headId", utils.FormatBlockId(head_id))
	bcmpbdsl.GetHeadToCheckpointChainRequest(*am.m, "bcm", fmt.Sprint(head_id), "application")
	return nil
}

func NewApplication(logger logging.Logger) modules.PassiveModule {

	m := dsl.NewModule("application")
	am := &ApplicationModule{m: &m, currentState: nil, logger: logger}

	dsl.UponInit(m, func() error {
		return nil
	})

	applicationpbdsl.UponPayloadRequest(m, am.handlePayloadRequest)
	applicationpbdsl.UponNewHead(m, am.handleNewHead)
	bcmpbdsl.UponGetHeadToCheckpointChainResponse(m, am.handleGetHeadToCheckpointChainResponse)

	return m
}
