package mscpbmsgs

import (
	types3 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func RequestSigMessage(destModule types.ModuleID, txs []*types1.Request, reqId uint64) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_MultisigCollector{
			MultisigCollector: &types3.Message{
				Type: &types3.Message_RequestSig{
					RequestSig: &types3.RequestSigMessage{
						Txs:   txs,
						ReqId: reqId,
					},
				},
			},
		},
	}
}

func SigMessage(destModule types.ModuleID, signature []uint8, reqId uint64) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_MultisigCollector{
			MultisigCollector: &types3.Message{
				Type: &types3.Message_Sig{
					Sig: &types3.SigMessage{
						Signature: signature,
						ReqId:     reqId,
					},
				},
			},
		},
	}
}

func RequestBatchMessage(destModule types.ModuleID, batchId []uint8, reqId uint64) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_MultisigCollector{
			MultisigCollector: &types3.Message{
				Type: &types3.Message_RequestBatch{
					RequestBatch: &types3.RequestBatchMessage{
						BatchId: batchId,
						ReqId:   reqId,
					},
				},
			},
		},
	}
}

func ProvideBatchMessage(destModule types.ModuleID, txs []*types1.Request, reqId uint64, batchId []uint8) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_MultisigCollector{
			MultisigCollector: &types3.Message{
				Type: &types3.Message_ProvideBatch{
					ProvideBatch: &types3.ProvideBatchMessage{
						Txs:     txs,
						ReqId:   reqId,
						BatchId: batchId,
					},
				},
			},
		},
	}
}
