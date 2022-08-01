package protobuf

import (
	apb "github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

func Message(moduleID t.ModuleID, msg *mscpb.Message) *messagepb.Message {
	return &messagepb.Message{
		DestModule: moduleID.Pb(),
		Type: &messagepb.Message_MultisigCollector{
			MultisigCollector: msg,
		},
	}
}

func RequestSigMessage(moduleID t.ModuleID, txs []*requestpb.Request, reqID uint64) *messagepb.Message {
	return Message(moduleID, &mscpb.Message{
		Type: &mscpb.Message_RequestSig{
			RequestSig: &mscpb.RequestSigMessage{
				Txs:   txs,
				ReqId: reqID,
			},
		},
	})
}

func SigMessage(moduleID t.ModuleID, signature []byte, reqID uint64) *messagepb.Message {
	return Message(moduleID, &mscpb.Message{
		Type: &mscpb.Message_Sig{
			Sig: &mscpb.SigMessage{
				Signature: signature,
				ReqId:     reqID,
			},
		},
	})
}

func RequestBatchMessage(moduleID t.ModuleID, batchID t.BatchID, reqID uint64) *messagepb.Message {
	return Message(moduleID, &mscpb.Message{
		Type: &mscpb.Message_RequestBatch{
			RequestBatch: &mscpb.RequestBatchMessage{
				BatchId: batchID.Pb(),
				ReqId:   reqID,
			},
		},
	})
}

func ProvideBatchMessage(moduleID t.ModuleID, txs []*requestpb.Request, reqID uint64) *messagepb.Message {
	return Message(moduleID, &mscpb.Message{
		Type: &mscpb.Message_ProvideBatch{
			ProvideBatch: &mscpb.ProvideBatchMessage{
				Txs:   txs,
				ReqId: reqID,
			},
		},
	})
}

func Cert(batchID t.BatchID, signers []t.NodeID, signatures [][]byte) *apb.Cert {
	return &apb.Cert{
		Type: &apb.Cert_Msc{
			Msc: &mscpb.Cert{
				BatchId:    batchID.Pb(),
				Signers:    t.NodeIDSlicePb(signers),
				Signatures: signatures,
			},
		},
	}
}
