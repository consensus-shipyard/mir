package batchfetcherpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	batchfetcherpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	trantorpb "github.com/filecoin-project/mir/pkg/pb/trantorpb"
	types "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() batchfetcherpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb batchfetcherpb.Event_Type) Event_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *batchfetcherpb.Event_NewOrderedBatch:
		return &Event_NewOrderedBatch{NewOrderedBatch: NewOrderedBatchFromPb(pb.NewOrderedBatch)}
	case *batchfetcherpb.Event_ClientProgress:
		return &Event_ClientProgress{ClientProgress: types.ClientProgressFromPb(pb.ClientProgress)}
	}
	return nil
}

type Event_NewOrderedBatch struct {
	NewOrderedBatch *NewOrderedBatch
}

func (*Event_NewOrderedBatch) isEvent_Type() {}

func (w *Event_NewOrderedBatch) Unwrap() *NewOrderedBatch {
	return w.NewOrderedBatch
}

func (w *Event_NewOrderedBatch) Pb() batchfetcherpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.NewOrderedBatch == nil {
		return &batchfetcherpb.Event_NewOrderedBatch{}
	}
	return &batchfetcherpb.Event_NewOrderedBatch{NewOrderedBatch: (w.NewOrderedBatch).Pb()}
}

func (*Event_NewOrderedBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchfetcherpb.Event_NewOrderedBatch]()}
}

type Event_ClientProgress struct {
	ClientProgress *types.ClientProgress
}

func (*Event_ClientProgress) isEvent_Type() {}

func (w *Event_ClientProgress) Unwrap() *types.ClientProgress {
	return w.ClientProgress
}

func (w *Event_ClientProgress) Pb() batchfetcherpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.ClientProgress == nil {
		return &batchfetcherpb.Event_ClientProgress{}
	}
	return &batchfetcherpb.Event_ClientProgress{ClientProgress: (w.ClientProgress).Pb()}
}

func (*Event_ClientProgress) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchfetcherpb.Event_ClientProgress]()}
}

func EventFromPb(pb *batchfetcherpb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *batchfetcherpb.Event {
	if m == nil {
		return nil
	}
	pbMessage := &batchfetcherpb.Event{}
	{
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchfetcherpb.Event]()}
}

type NewOrderedBatch struct {
	Txs []*types.Transaction
}

func NewOrderedBatchFromPb(pb *batchfetcherpb.NewOrderedBatch) *NewOrderedBatch {
	if pb == nil {
		return nil
	}
	return &NewOrderedBatch{
		Txs: types1.ConvertSlice(pb.Txs, func(t *trantorpb.Transaction) *types.Transaction {
			return types.TransactionFromPb(t)
		}),
	}
}

func (m *NewOrderedBatch) Pb() *batchfetcherpb.NewOrderedBatch {
	if m == nil {
		return nil
	}
	pbMessage := &batchfetcherpb.NewOrderedBatch{}
	{
		pbMessage.Txs = types1.ConvertSlice(m.Txs, func(t *types.Transaction) *trantorpb.Transaction {
			return (t).Pb()
		})
	}

	return pbMessage
}

func (*NewOrderedBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchfetcherpb.NewOrderedBatch]()}
}
