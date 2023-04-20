package hasherpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	types "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	hasherpb "github.com/filecoin-project/mir/pkg/pb/hasherpb"
	ordererpb "github.com/filecoin-project/mir/pkg/pb/ordererpb"
	types4 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() hasherpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb hasherpb.Event_Type) Event_Type {
	switch pb := pb.(type) {
	case *hasherpb.Event_Request:
		return &Event_Request{Request: RequestFromPb(pb.Request)}
	case *hasherpb.Event_Result:
		return &Event_Result{Result: ResultFromPb(pb.Result)}
	case *hasherpb.Event_RequestOne:
		return &Event_RequestOne{RequestOne: RequestOneFromPb(pb.RequestOne)}
	case *hasherpb.Event_ResultOne:
		return &Event_ResultOne{ResultOne: ResultOneFromPb(pb.ResultOne)}
	}
	return nil
}

type Event_Request struct {
	Request *Request
}

func (*Event_Request) isEvent_Type() {}

func (w *Event_Request) Unwrap() *Request {
	return w.Request
}

func (w *Event_Request) Pb() hasherpb.Event_Type {
	return &hasherpb.Event_Request{Request: (w.Request).Pb()}
}

func (*Event_Request) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Event_Request]()}
}

type Event_Result struct {
	Result *Result
}

func (*Event_Result) isEvent_Type() {}

func (w *Event_Result) Unwrap() *Result {
	return w.Result
}

func (w *Event_Result) Pb() hasherpb.Event_Type {
	return &hasherpb.Event_Result{Result: (w.Result).Pb()}
}

func (*Event_Result) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Event_Result]()}
}

type Event_RequestOne struct {
	RequestOne *RequestOne
}

func (*Event_RequestOne) isEvent_Type() {}

func (w *Event_RequestOne) Unwrap() *RequestOne {
	return w.RequestOne
}

func (w *Event_RequestOne) Pb() hasherpb.Event_Type {
	return &hasherpb.Event_RequestOne{RequestOne: (w.RequestOne).Pb()}
}

func (*Event_RequestOne) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Event_RequestOne]()}
}

type Event_ResultOne struct {
	ResultOne *ResultOne
}

func (*Event_ResultOne) isEvent_Type() {}

func (w *Event_ResultOne) Unwrap() *ResultOne {
	return w.ResultOne
}

func (w *Event_ResultOne) Pb() hasherpb.Event_Type {
	return &hasherpb.Event_ResultOne{ResultOne: (w.ResultOne).Pb()}
}

func (*Event_ResultOne) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Event_ResultOne]()}
}

func EventFromPb(pb *hasherpb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *hasherpb.Event {
	return &hasherpb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Event]()}
}

type Request struct {
	Data   []*types.HashData
	Origin *HashOrigin
}

func RequestFromPb(pb *hasherpb.Request) *Request {
	return &Request{
		Data: types1.ConvertSlice(pb.Data, func(t *commonpb.HashData) *types.HashData {
			return types.HashDataFromPb(t)
		}),
		Origin: HashOriginFromPb(pb.Origin),
	}
}

func (m *Request) Pb() *hasherpb.Request {
	return &hasherpb.Request{
		Data: types1.ConvertSlice(m.Data, func(t *types.HashData) *commonpb.HashData {
			return (t).Pb()
		}),
		Origin: (m.Origin).Pb(),
	}
}

func (*Request) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Request]()}
}

type Result struct {
	Digests [][]uint8
	Origin  *HashOrigin
}

func ResultFromPb(pb *hasherpb.Result) *Result {
	return &Result{
		Digests: pb.Digests,
		Origin:  HashOriginFromPb(pb.Origin),
	}
}

func (m *Result) Pb() *hasherpb.Result {
	return &hasherpb.Result{
		Digests: m.Digests,
		Origin:  (m.Origin).Pb(),
	}
}

func (*Result) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Result]()}
}

type RequestOne struct {
	Data   *types.HashData
	Origin *HashOrigin
}

func RequestOneFromPb(pb *hasherpb.RequestOne) *RequestOne {
	return &RequestOne{
		Data:   types.HashDataFromPb(pb.Data),
		Origin: HashOriginFromPb(pb.Origin),
	}
}

func (m *RequestOne) Pb() *hasherpb.RequestOne {
	return &hasherpb.RequestOne{
		Data:   (m.Data).Pb(),
		Origin: (m.Origin).Pb(),
	}
}

func (*RequestOne) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.RequestOne]()}
}

type ResultOne struct {
	Digest []uint8
	Origin *HashOrigin
}

func ResultOneFromPb(pb *hasherpb.ResultOne) *ResultOne {
	return &ResultOne{
		Digest: pb.Digest,
		Origin: HashOriginFromPb(pb.Origin),
	}
}

func (m *ResultOne) Pb() *hasherpb.ResultOne {
	return &hasherpb.ResultOne{
		Digest: m.Digest,
		Origin: (m.Origin).Pb(),
	}
}

func (*ResultOne) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.ResultOne]()}
}

type HashOrigin struct {
	Module types2.ModuleID
	Type   HashOrigin_Type
}

type HashOrigin_Type interface {
	mirreflect.GeneratedType
	isHashOrigin_Type()
	Pb() hasherpb.HashOrigin_Type
}

type HashOrigin_TypeWrapper[T any] interface {
	HashOrigin_Type
	Unwrap() *T
}

func HashOrigin_TypeFromPb(pb hasherpb.HashOrigin_Type) HashOrigin_Type {
	switch pb := pb.(type) {
	case *hasherpb.HashOrigin_ContextStore:
		return &HashOrigin_ContextStore{ContextStore: types3.OriginFromPb(pb.ContextStore)}
	case *hasherpb.HashOrigin_Request:
		return &HashOrigin_Request{Request: types4.RequestFromPb(pb.Request)}
	case *hasherpb.HashOrigin_Dsl:
		return &HashOrigin_Dsl{Dsl: types5.OriginFromPb(pb.Dsl)}
	case *hasherpb.HashOrigin_Checkpoint:
		return &HashOrigin_Checkpoint{Checkpoint: types6.HashOriginFromPb(pb.Checkpoint)}
	case *hasherpb.HashOrigin_Sb:
		return &HashOrigin_Sb{Sb: pb.Sb}
	}
	return nil
}

type HashOrigin_ContextStore struct {
	ContextStore *types3.Origin
}

func (*HashOrigin_ContextStore) isHashOrigin_Type() {}

func (w *HashOrigin_ContextStore) Unwrap() *types3.Origin {
	return w.ContextStore
}

func (w *HashOrigin_ContextStore) Pb() hasherpb.HashOrigin_Type {
	return &hasherpb.HashOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*HashOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashOrigin_ContextStore]()}
}

type HashOrigin_Request struct {
	Request *types4.Request
}

func (*HashOrigin_Request) isHashOrigin_Type() {}

func (w *HashOrigin_Request) Unwrap() *types4.Request {
	return w.Request
}

func (w *HashOrigin_Request) Pb() hasherpb.HashOrigin_Type {
	return &hasherpb.HashOrigin_Request{Request: (w.Request).Pb()}
}

func (*HashOrigin_Request) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashOrigin_Request]()}
}

type HashOrigin_Dsl struct {
	Dsl *types5.Origin
}

func (*HashOrigin_Dsl) isHashOrigin_Type() {}

func (w *HashOrigin_Dsl) Unwrap() *types5.Origin {
	return w.Dsl
}

func (w *HashOrigin_Dsl) Pb() hasherpb.HashOrigin_Type {
	return &hasherpb.HashOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*HashOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashOrigin_Dsl]()}
}

type HashOrigin_Checkpoint struct {
	Checkpoint *types6.HashOrigin
}

func (*HashOrigin_Checkpoint) isHashOrigin_Type() {}

func (w *HashOrigin_Checkpoint) Unwrap() *types6.HashOrigin {
	return w.Checkpoint
}

func (w *HashOrigin_Checkpoint) Pb() hasherpb.HashOrigin_Type {
	return &hasherpb.HashOrigin_Checkpoint{Checkpoint: (w.Checkpoint).Pb()}
}

func (*HashOrigin_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashOrigin_Checkpoint]()}
}

type HashOrigin_Sb struct {
	Sb *ordererpb.HashOrigin
}

func (*HashOrigin_Sb) isHashOrigin_Type() {}

func (w *HashOrigin_Sb) Unwrap() *ordererpb.HashOrigin {
	return w.Sb
}

func (w *HashOrigin_Sb) Pb() hasherpb.HashOrigin_Type {
	return &hasherpb.HashOrigin_Sb{Sb: w.Sb}
}

func (*HashOrigin_Sb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashOrigin_Sb]()}
}

func HashOriginFromPb(pb *hasherpb.HashOrigin) *HashOrigin {
	return &HashOrigin{
		Module: (types2.ModuleID)(pb.Module),
		Type:   HashOrigin_TypeFromPb(pb.Type),
	}
}

func (m *HashOrigin) Pb() *hasherpb.HashOrigin {
	return &hasherpb.HashOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*HashOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashOrigin]()}
}
