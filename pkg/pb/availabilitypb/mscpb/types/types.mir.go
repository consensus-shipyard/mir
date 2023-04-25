package mscpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	mscpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	trantorpb "github.com/filecoin-project/mir/pkg/pb/trantorpb"
	types "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
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
	Txs   []*types.Transaction
	ReqId uint64
}

func RequestSigMessageFromPb(pb *mscpb.RequestSigMessage) *RequestSigMessage {
	return &RequestSigMessage{
		Txs: types1.ConvertSlice(pb.Txs, func(t *trantorpb.Transaction) *types.Transaction {
			return types.TransactionFromPb(t)
		}),
		ReqId: pb.ReqId,
	}
}

func (m *RequestSigMessage) Pb() *mscpb.RequestSigMessage {
	return &mscpb.RequestSigMessage{
		Txs: types1.ConvertSlice(m.Txs, func(t *types.Transaction) *trantorpb.Transaction {
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
	ReqId   uint64
}

func RequestBatchMessageFromPb(pb *mscpb.RequestBatchMessage) *RequestBatchMessage {
	return &RequestBatchMessage{
		BatchId: pb.BatchId,
		ReqId:   pb.ReqId,
	}
}

func (m *RequestBatchMessage) Pb() *mscpb.RequestBatchMessage {
	return &mscpb.RequestBatchMessage{
		BatchId: m.BatchId,
		ReqId:   m.ReqId,
	}
}

func (*RequestBatchMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.RequestBatchMessage]()}
}

type ProvideBatchMessage struct {
	Txs     []*types.Transaction
	ReqId   uint64
	BatchId []uint8
}

func ProvideBatchMessageFromPb(pb *mscpb.ProvideBatchMessage) *ProvideBatchMessage {
	return &ProvideBatchMessage{
		Txs: types1.ConvertSlice(pb.Txs, func(t *trantorpb.Transaction) *types.Transaction {
			return types.TransactionFromPb(t)
		}),
		ReqId:   pb.ReqId,
		BatchId: pb.BatchId,
	}
}

func (m *ProvideBatchMessage) Pb() *mscpb.ProvideBatchMessage {
	return &mscpb.ProvideBatchMessage{
		Txs: types1.ConvertSlice(m.Txs, func(t *types.Transaction) *trantorpb.Transaction {
			return (t).Pb()
		}),
		ReqId:   m.ReqId,
		BatchId: m.BatchId,
	}
}

func (*ProvideBatchMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.ProvideBatchMessage]()}
}

type Cert struct {
	BatchId    []uint8
	Signers    []types2.NodeID
	Signatures [][]uint8
}

func CertFromPb(pb *mscpb.Cert) *Cert {
	return &Cert{
		BatchId: pb.BatchId,
		Signers: types1.ConvertSlice(pb.Signers, func(t string) types2.NodeID {
			return (types2.NodeID)(t)
		}),
		Signatures: pb.Signatures,
	}
}

func (m *Cert) Pb() *mscpb.Cert {
	return &mscpb.Cert{
		BatchId: m.BatchId,
		Signers: types1.ConvertSlice(m.Signers, func(t types2.NodeID) string {
			return (string)(t)
		}),
		Signatures: m.Signatures,
	}
}

func (*Cert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.Cert]()}
}

type Certs struct {
	Certs []*Cert
}

func CertsFromPb(pb *mscpb.Certs) *Certs {
	return &Certs{
		Certs: types1.ConvertSlice(pb.Certs, func(t *mscpb.Cert) *Cert {
			return CertFromPb(t)
		}),
	}
}

func (m *Certs) Pb() *mscpb.Certs {
	return &mscpb.Certs{
		Certs: types1.ConvertSlice(m.Certs, func(t *Cert) *mscpb.Cert {
			return (t).Pb()
		}),
	}
}

func (*Certs) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.Certs]()}
}

type InstanceParams struct {
	Membership  *types.Membership
	Limit       uint64
	MaxRequests uint64
}

func InstanceParamsFromPb(pb *mscpb.InstanceParams) *InstanceParams {
	return &InstanceParams{
		Membership:  types.MembershipFromPb(pb.Membership),
		Limit:       pb.Limit,
		MaxRequests: pb.MaxRequests,
	}
}

func (m *InstanceParams) Pb() *mscpb.InstanceParams {
	return &mscpb.InstanceParams{
		Membership:  (m.Membership).Pb(),
		Limit:       m.Limit,
		MaxRequests: m.MaxRequests,
	}
}

func (*InstanceParams) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mscpb.InstanceParams]()}
}
