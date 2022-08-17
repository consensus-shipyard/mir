package batchfetcherpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	batchfetcherpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
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
	switch pb := pb.(type) {
	case *batchfetcherpb.Event_NewOrderedBatch:
		return &Event_NewOrderedBatch{NewOrderedBatch: NewOrderedBatchFromPb(pb.NewOrderedBatch)}
	case *batchfetcherpb.Event_ClientProgress:
		return &Event_ClientProgress{ClientProgress: pb.ClientProgress}
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
	return &batchfetcherpb.Event_NewOrderedBatch{NewOrderedBatch: (w.NewOrderedBatch).Pb()}
}

func (*Event_NewOrderedBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchfetcherpb.Event_NewOrderedBatch]()}
}

type Event_ClientProgress struct {
	ClientProgress *commonpb.ClientProgress
}

func (*Event_ClientProgress) isEvent_Type() {}

func (w *Event_ClientProgress) Unwrap() *commonpb.ClientProgress {
	return w.ClientProgress
}

func (w *Event_ClientProgress) Pb() batchfetcherpb.Event_Type {
	return &batchfetcherpb.Event_ClientProgress{ClientProgress: w.ClientProgress}
}

func (*Event_ClientProgress) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchfetcherpb.Event_ClientProgress]()}
}

func EventFromPb(pb *batchfetcherpb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *batchfetcherpb.Event {
	return &batchfetcherpb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchfetcherpb.Event]()}
}

type NewOrderedBatch struct {
	Txs []*requestpb.Request
}

func NewOrderedBatchFromPb(pb *batchfetcherpb.NewOrderedBatch) *NewOrderedBatch {
	return &NewOrderedBatch{
		Txs: pb.Txs,
	}
}

func (m *NewOrderedBatch) Pb() *batchfetcherpb.NewOrderedBatch {
	return &batchfetcherpb.NewOrderedBatch{
		Txs: m.Txs,
	}
}

func (*NewOrderedBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchfetcherpb.NewOrderedBatch]()}
}
