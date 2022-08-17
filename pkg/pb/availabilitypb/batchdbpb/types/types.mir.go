package batchdbpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	batchdbpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() batchdbpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb batchdbpb.Event_Type) Event_Type {
	switch pb := pb.(type) {
	case *batchdbpb.Event_Lookup:
		return &Event_Lookup{Lookup: pb.Lookup}
	case *batchdbpb.Event_LookupResponse:
		return &Event_LookupResponse{LookupResponse: pb.LookupResponse}
	case *batchdbpb.Event_Store:
		return &Event_Store{Store: pb.Store}
	case *batchdbpb.Event_Stored:
		return &Event_Stored{Stored: pb.Stored}
	}
	return nil
}

type Event_Lookup struct {
	Lookup *batchdbpb.LookupBatch
}

func (*Event_Lookup) isEvent_Type() {}

func (w *Event_Lookup) Unwrap() *batchdbpb.LookupBatch {
	return w.Lookup
}

func (w *Event_Lookup) Pb() batchdbpb.Event_Type {
	return &batchdbpb.Event_Lookup{Lookup: w.Lookup}
}

func (*Event_Lookup) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event_Lookup]()}
}

type Event_LookupResponse struct {
	LookupResponse *batchdbpb.LookupBatchResponse
}

func (*Event_LookupResponse) isEvent_Type() {}

func (w *Event_LookupResponse) Unwrap() *batchdbpb.LookupBatchResponse {
	return w.LookupResponse
}

func (w *Event_LookupResponse) Pb() batchdbpb.Event_Type {
	return &batchdbpb.Event_LookupResponse{LookupResponse: w.LookupResponse}
}

func (*Event_LookupResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event_LookupResponse]()}
}

type Event_Store struct {
	Store *batchdbpb.StoreBatch
}

func (*Event_Store) isEvent_Type() {}

func (w *Event_Store) Unwrap() *batchdbpb.StoreBatch {
	return w.Store
}

func (w *Event_Store) Pb() batchdbpb.Event_Type {
	return &batchdbpb.Event_Store{Store: w.Store}
}

func (*Event_Store) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event_Store]()}
}

type Event_Stored struct {
	Stored *batchdbpb.BatchStored
}

func (*Event_Stored) isEvent_Type() {}

func (w *Event_Stored) Unwrap() *batchdbpb.BatchStored {
	return w.Stored
}

func (w *Event_Stored) Pb() batchdbpb.Event_Type {
	return &batchdbpb.Event_Stored{Stored: w.Stored}
}

func (*Event_Stored) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event_Stored]()}
}

func EventFromPb(pb *batchdbpb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *batchdbpb.Event {
	return &batchdbpb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event]()}
}
