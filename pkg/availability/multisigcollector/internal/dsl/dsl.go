package mscdsl

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/pb/requestpb"

	adsl "github.com/filecoin-project/mir/pkg/availability/dsl"
	"github.com/filecoin-project/mir/pkg/dsl"
	apb "github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponMscMessageReceived(m dsl.Module, handler func(from t.NodeID, msg *mscpb.Message) error) {
	dsl.UponMessageReceived(m, func(from t.NodeID, msg *messagepb.Message) error {
		cbMsgWrapper, ok := msg.Type.(*messagepb.Message_MultisigCollector)
		if !ok {
			return nil
		}

		return handler(from, cbMsgWrapper.MultisigCollector)
	})
}

func UponRequestSigMessageReceived(m dsl.Module, handler func(from t.NodeID, txs []*requestpb.Request, reqID uint64) error) {
	UponMscMessageReceived(m, func(from t.NodeID, msg *mscpb.Message) error {
		requestSigMsgWrapper, ok := msg.Type.(*mscpb.Message_RequestSig)
		if !ok {
			return nil
		}
		requestSigMsg := requestSigMsgWrapper.RequestSig

		return handler(from, requestSigMsg.Txs, requestSigMsg.ReqId)
	})
}

func UponSigMessageReceived(m dsl.Module, handler func(from t.NodeID, signature []byte, reqID uint64) error) {
	UponMscMessageReceived(m, func(from t.NodeID, msg *mscpb.Message) error {
		sigMsgWrapper, ok := msg.Type.(*mscpb.Message_Sig)
		if !ok {
			return nil
		}
		sigMsg := sigMsgWrapper.Sig

		return handler(from, sigMsg.Signature, sigMsg.ReqId)
	})
}

func UponRequestBatchMessageReceived(m dsl.Module, handler func(from t.NodeID, batchID t.BatchID, reqID uint64) error) {
	UponMscMessageReceived(m, func(from t.NodeID, msg *mscpb.Message) error {
		requestBatchMsgWrapper, ok := msg.Type.(*mscpb.Message_RequestBatch)
		if !ok {
			return nil
		}
		requestBatchMsg := requestBatchMsgWrapper.RequestBatch

		return handler(from, t.BatchID(requestBatchMsg.BatchId), requestBatchMsg.ReqId)
	})
}

func UponProvideBatchMessageReceived(m dsl.Module, handler func(from t.NodeID, txs []*requestpb.Request, reqID uint64) error) {
	UponMscMessageReceived(m, func(from t.NodeID, msg *mscpb.Message) error {
		provideBatchMsgWrapper, ok := msg.Type.(*mscpb.Message_ProvideBatch)
		if !ok {
			return nil
		}
		provideBatchMsg := provideBatchMsgWrapper.ProvideBatch

		return handler(from, provideBatchMsg.Txs, provideBatchMsg.ReqId)
	})
}

func UponRequestTransactions(m dsl.Module, handler func(cert *mscpb.Cert, origin *apb.RequestTransactionsOrigin) error) {
	adsl.UponRequestTransactions(m, func(cert *apb.Cert, origin *apb.RequestTransactionsOrigin) error {
		mscCertWrapper, ok := cert.Type.(*apb.Cert_Msc)
		if !ok {
			return fmt.Errorf("unexpected certificate type. Expected: %T, got: %T", mscCertWrapper, cert.Type)
		}

		return handler(mscCertWrapper.Msc, origin)
	})
}
