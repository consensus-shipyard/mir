// Code generated by Mir codegen. DO NOT EDIT.

package bcmpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types3 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Bcm](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponNewBlock(m dsl.Module, handler func(block *types2.Block) error) {
	UponEvent[*types.Event_NewBlock](m, func(ev *types.NewBlock) error {
		return handler(ev.Block)
	})
}

func UponNewChain(m dsl.Module, handler func(blocks []*types2.Block) error) {
	UponEvent[*types.Event_NewChain](m, func(ev *types.NewChain) error {
		return handler(ev.Blocks)
	})
}

func UponGetBlockRequest(m dsl.Module, handler func(requestId string, sourceModule types3.ModuleID, blockId uint64) error) {
	UponEvent[*types.Event_GetBlockRequest](m, func(ev *types.GetBlockRequest) error {
		return handler(ev.RequestId, ev.SourceModule, ev.BlockId)
	})
}

func UponGetBlockResponse(m dsl.Module, handler func(requestId string, found bool, block *types2.Block) error) {
	UponEvent[*types.Event_GetBlockResponse](m, func(ev *types.GetBlockResponse) error {
		return handler(ev.RequestId, ev.Found, ev.Block)
	})
}

func UponGetChainRequest(m dsl.Module, handler func(requestId string, sourceModule types3.ModuleID, endBlockId uint64, sourceBlockIds []uint64) error) {
	UponEvent[*types.Event_GetChainRequest](m, func(ev *types.GetChainRequest) error {
		return handler(ev.RequestId, ev.SourceModule, ev.EndBlockId, ev.SourceBlockIds)
	})
}

func UponGetChainResponse(m dsl.Module, handler func(requestId string, success bool, chain []*types2.Block) error) {
	UponEvent[*types.Event_GetChainResponse](m, func(ev *types.GetChainResponse) error {
		return handler(ev.RequestId, ev.Success, ev.Chain)
	})
}

func UponGetHeadToCheckpointChainRequest(m dsl.Module, handler func(requestId string, sourceModule types3.ModuleID) error) {
	UponEvent[*types.Event_GetHeadToCheckpointChainRequest](m, func(ev *types.GetHeadToCheckpointChainRequest) error {
		return handler(ev.RequestId, ev.SourceModule)
	})
}

func UponGetHeadToCheckpointChainResponse(m dsl.Module, handler func(requestId string, chain []*types2.Block, checkpointState *types4.State) error) {
	UponEvent[*types.Event_GetHeadToCheckpointChainResponse](m, func(ev *types.GetHeadToCheckpointChainResponse) error {
		return handler(ev.RequestId, ev.Chain, ev.CheckpointState)
	})
}

func UponRegisterCheckpoint(m dsl.Module, handler func(blockId uint64, state *types4.State) error) {
	UponEvent[*types.Event_RegisterCheckpoint](m, func(ev *types.RegisterCheckpoint) error {
		return handler(ev.BlockId, ev.State)
	})
}

func UponInitBlockchain(m dsl.Module, handler func(initialState *types4.State) error) {
	UponEvent[*types.Event_InitBlockchain](m, func(ev *types.InitBlockchain) error {
		return handler(ev.InitialState)
	})
}
