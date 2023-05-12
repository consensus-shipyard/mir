package mscpbdsl

import (
	types4 "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/messagepb/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing net messages.

func UponMessageReceived[W types.Message_TypeWrapper[M], M any](m dsl.Module, handler func(from types1.NodeID, msg *M) error) {
	dsl1.UponMessageReceived[*types2.Message_MultisigCollector](m, func(from types1.NodeID, msg *types.Message) error {
		w, ok := msg.Type.(W)
		if !ok {
			return nil
		}

		return handler(from, w.Unwrap())
	})
}

func UponRequestSigMessageReceived(m dsl.Module, handler func(from types1.NodeID, txs []*types3.Transaction, reqId uint64) error) {
	UponMessageReceived[*types.Message_RequestSig](m, func(from types1.NodeID, msg *types.RequestSigMessage) error {
		return handler(from, msg.Txs, msg.ReqId)
	})
}

func UponSigMessageReceived(m dsl.Module, handler func(from types1.NodeID, signature []uint8, reqId uint64) error) {
	UponMessageReceived[*types.Message_Sig](m, func(from types1.NodeID, msg *types.SigMessage) error {
		return handler(from, msg.Signature, msg.ReqId)
	})
}

func UponRequestBatchMessageReceived(m dsl.Module, handler func(from types1.NodeID, batchId types4.BatchID, reqId uint64) error) {
	UponMessageReceived[*types.Message_RequestBatch](m, func(from types1.NodeID, msg *types.RequestBatchMessage) error {
		return handler(from, msg.BatchId, msg.ReqId)
	})
}

func UponProvideBatchMessageReceived(m dsl.Module, handler func(from types1.NodeID, txs []*types3.Transaction, reqId uint64, batchId types4.BatchID) error) {
	UponMessageReceived[*types.Message_ProvideBatch](m, func(from types1.NodeID, msg *types.ProvideBatchMessage) error {
		return handler(from, msg.Txs, msg.ReqId, msg.BatchId)
	})
}
