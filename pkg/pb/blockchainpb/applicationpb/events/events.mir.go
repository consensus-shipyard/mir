// Code generated by Mir codegen. DO NOT EDIT.

package applicationpbevents

import (
	blockchainpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	types2 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/types"
	payloadpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb"
	statepb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func NewHead(destModule types.ModuleID, headId uint64) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Application{
			Application: &types2.Event{
				Type: &types2.Event_NewHead{
					NewHead: &types2.NewHead{
						HeadId: headId,
					},
				},
			},
		},
	}
}

func VerifyBlockRequest(destModule types.ModuleID, requestId uint64, block *blockchainpb.Block) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Application{
			Application: &types2.Event{
				Type: &types2.Event_VerifyBlockRequest{
					VerifyBlockRequest: &types2.VerifyBlockRequest{
						RequestId: requestId,
						Block:     block,
					},
				},
			},
		},
	}
}

func VerifyBlockResponse(destModule types.ModuleID, requestId uint64, ok bool) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Application{
			Application: &types2.Event{
				Type: &types2.Event_VerifyBlockResponse{
					VerifyBlockResponse: &types2.VerifyBlockResponse{
						RequestId: requestId,
						Ok:        ok,
					},
				},
			},
		},
	}
}

func PayloadRequest(destModule types.ModuleID, headId uint64) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Application{
			Application: &types2.Event{
				Type: &types2.Event_PayloadRequest{
					PayloadRequest: &types2.PayloadRequest{
						HeadId: headId,
					},
				},
			},
		},
	}
}

func PayloadResponse(destModule types.ModuleID, headId uint64, payload *payloadpb.Payload) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Application{
			Application: &types2.Event{
				Type: &types2.Event_PayloadResponse{
					PayloadResponse: &types2.PayloadResponse{
						HeadId:  headId,
						Payload: payload,
					},
				},
			},
		},
	}
}

func ForkUpdate(destModule types.ModuleID, removedChain *blockchainpb.Blockchain, addedChain *blockchainpb.Blockchain, forkState *statepb.State) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Application{
			Application: &types2.Event{
				Type: &types2.Event_ForkUpdate{
					ForkUpdate: &types2.ForkUpdate{
						RemovedChain: removedChain,
						AddedChain:   addedChain,
						ForkState:    forkState,
					},
				},
			},
		},
	}
}
