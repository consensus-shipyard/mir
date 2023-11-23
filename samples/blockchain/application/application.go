package application

// app module

import (
	"math/rand"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	applicationpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/dsl"
	interceptorpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/interceptorpb/dsl"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb"
	"github.com/filecoin-project/mir/samples/blockchain/utils"
)

type ApplicationModule struct {
	m      *dsl.Module
	state  *state
	logger logging.Logger
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

func (am *ApplicationModule) handleRegisterBlock(block *blockchainpb.Block) error {
	am.state.applyBlock(block)
	am.logger.Log(logging.LevelDebug, "Registered new block", "block_id", utils.FormatBlockId(block.BlockId))
	return nil
}

func (am *ApplicationModule) handleNewHead(head_id uint64) error {
	previousState := am.state.getCurrentState()
	am.state.setHead(head_id)
	currentState := am.state.getCurrentState()
	interceptorpbdsl.AppUpdate(*am.m, "devnull", currentState)
	am.logger.Log(logging.LevelInfo, "Application state updated", "previousState", previousState, "currentState", currentState, "headId", utils.FormatBlockId(head_id))
	return nil
}

func NewApplication(logger logging.Logger) modules.PassiveModule {

	m := dsl.NewModule("application")
	am := &ApplicationModule{m: &m, state: initState(), logger: logger}

	dsl.UponInit(m, func() error {
		return nil
	})

	applicationpbdsl.UponPayloadRequest(m, am.handlePayloadRequest)
	applicationpbdsl.UponNewHead(m, am.handleNewHead)
	applicationpbdsl.UponRegisterBlock(m, am.handleRegisterBlock)

	return m
}
