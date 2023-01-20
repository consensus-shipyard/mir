package mscpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	mscpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Message struct {
	Type Message_Type
}

type Message_Type interface {
	mirreflect.GeneratedType
	isMessage_Type()
	Pb() mscpb.Message_Type
}

type Message_TypeWrapper[T any] interface {
	Message_Type
	Unwrap() *T
}

func Message_TypeFromPb(pb mscpb.Message_Type) Message_Type {
	switch pb := pb.(type) {
	case *mscpb.Message_RequestSig:
		return &Message_RequestSig{RequestSig: RequestSigMessageFromPb(pb.RequestSig)}
	case *mscpb.Message_Sig:
		return &Message_Sig{Sig: SigMessageFromPb(pb.Sig)}
	case *mscpb.Message_RequestBatch:
		return &Message_RequestBatch{RequestBatch: RequestBatchMessageFromPb(pb.RequestBatch)}
	case *mscpb.Message_ProvideBatch:
		return &Message_ProvideBatch{ProvideBatch: ProvideBatchMessageFromPb(pb.ProvideBatch)}
	}
	return nil
}

type Message_RequestSig struct {
	RequestSig *RequestSigMessage
}

func (*Message_RequestSig) isMessage_Type() {}

func (w *Message_RequestSig) Unwrap() *RequestSigMessage {
	return w.RequestSig
}

func (w *Message_RequestSig) Pb() mscpb.Message_Type {
	return &mscpb.Message_RequestSig{RequestSig: (w.RequestSig).Pb()}
}

func (*Message_RequestSig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.Message_RequestSig]()}
}

type Message_Sig struct {
	Sig *SigMessage
}

func (*Message_Sig) isMessage_Type() {}

func (w *Message_Sig) Unwrap() *SigMessage {
	return w.Sig
}

func (w *Message_Sig) Pb() mscpb.Message_Type {
	return &mscpb.Message_Sig{Sig: (w.Sig).Pb()}
}

func (*Message_Sig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.Message_Sig]()}
}

type Message_RequestBatch struct {
	RequestBatch *RequestBatchMessage
}

func (*Message_RequestBatch) isMessage_Type() {}

func (w *Message_RequestBatch) Unwrap() *RequestBatchMessage {
	return w.RequestBatch
}

func (w *Message_RequestBatch) Pb() mscpb.Message_Type {
	return &mscpb.Message_RequestBatch{RequestBatch: (w.RequestBatch).Pb()}
}

func (*Message_RequestBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.Message_RequestBatch]()}
}

type Message_ProvideBatch struct {
	ProvideBatch *ProvideBatchMessage
}

func (*Message_ProvideBatch) isMessage_Type() {}

func (w *Message_ProvideBatch) Unwrap() *ProvideBatchMessage {
	return w.ProvideBatch
}

func (w *Message_ProvideBatch) Pb() mscpb.Message_Type {
	return &mscpb.Message_ProvideBatch{ProvideBatch: (w.ProvideBatch).Pb()}
}

func (*Message_ProvideBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.Message_ProvideBatch]()}
}

func MessageFromPb(pb *mscpb.Message) *Message {
	return &Message{
		Type: Message_TypeFromPb(pb.Type),
	}
}

func (m *Message) Pb() *mscpb.Message {
	return &mscpb.Message{
		Type: (m.Type).Pb(),
	}
}

func (*Message) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.Message]()}
}

type RequestSigMessage struct {
	Txs   []*types.Request
	ReqId uint64
}

func RequestSigMessageFromPb(pb *mscpb.RequestSigMessage) *RequestSigMessage {
	return &RequestSigMessage{
		Txs: types1.ConvertSlice(pb.Txs, func(t *requestpb.Request) *types.Request {
			return types.RequestFromPb(t)
		}),
		ReqId: pb.ReqId,
	}
}

func (m *RequestSigMessage) Pb() *mscpb.RequestSigMessage {
	return &mscpb.RequestSigMessage{
		Txs: types1.ConvertSlice(m.Txs, func(t *types.Request) *requestpb.Request {
			return (t).Pb()
		}),
		ReqId: m.ReqId,
	}
}

func (*RequestSigMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.RequestSigMessage]()}
}

type SigMessage struct {
	Signature []uint8
	ReqId     uint64
}

func SigMessageFromPb(pb *mscpb.SigMessage) *SigMessage {
	return &SigMessage{
		Signature: pb.Signature,
		ReqId:     pb.ReqId,
	}
}

func (m *SigMessage) Pb() *mscpb.SigMessage {
	return &mscpb.SigMessage{
		Signature: m.Signature,
		ReqId:     m.ReqId,
	}
}

func (*SigMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.SigMessage]()}
}

type RequestBatchMessage struct {
	BatchId []uint8
	ReqId   types2.RequestID
}

func RequestBatchMessageFromPb(pb *mscpb.RequestBatchMessage) *RequestBatchMessage {
	return &RequestBatchMessage{
		BatchId: pb.BatchId,
		ReqId:   (types2.RequestID)(pb.ReqId),
	}
}

func (m *RequestBatchMessage) Pb() *mscpb.RequestBatchMessage {
	return &mscpb.RequestBatchMessage{
		BatchId: m.BatchId,
		ReqId:   (uint64)(m.ReqId),
	}
}

func (*RequestBatchMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.RequestBatchMessage]()}
}

type ProvideBatchMessage struct {
	Txs   []*types.Request
	ReqId uint64
}

func ProvideBatchMessageFromPb(pb *mscpb.ProvideBatchMessage) *ProvideBatchMessage {
	return &ProvideBatchMessage{
		Txs: types1.ConvertSlice(pb.Txs, func(t *requestpb.Request) *types.Request {
			return types.RequestFromPb(t)
		}),
		ReqId: pb.ReqId,
	}
}

func (m *ProvideBatchMessage) Pb() *mscpb.ProvideBatchMessage {
	return &mscpb.ProvideBatchMessage{
		Txs: types1.ConvertSlice(m.Txs, func(t *types.Request) *requestpb.Request {
			return (t).Pb()
		}),
		ReqId: m.ReqId,
	}
}

func (*ProvideBatchMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.ProvideBatchMessage]()}
}

type Cert struct {
	BatchId    []uint8
	Signers    []string
	Signatures [][]uint8
}

func CertFromPb(pb *mscpb.Cert) *Cert {
	return &Cert{
		BatchId:    pb.BatchId,
		Signers:    pb.Signers,
		Signatures: pb.Signatures,
	}
}

func (m *Cert) Pb() *mscpb.Cert {
	return &mscpb.Cert{
		BatchId:    m.BatchId,
		Signers:    m.Signers,
		Signatures: m.Signatures,
	}
}

func (*Cert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.Cert]()}
}
