package hasherpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types "github.com/filecoin-project/mir/codegen/model/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	hasherpb "github.com/filecoin-project/mir/pkg/pb/hasherpb"
	types1 "github.com/filecoin-project/mir/pkg/types"
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
	if pb == nil {
		return nil
	}
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
	if w == nil {
		return nil
	}
	if w.Request == nil {
		return &hasherpb.Event_Request{}
	}
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
	if w == nil {
		return nil
	}
	if w.Result == nil {
		return &hasherpb.Event_Result{}
	}
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
	if w == nil {
		return nil
	}
	if w.RequestOne == nil {
		return &hasherpb.Event_RequestOne{}
	}
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
	if w == nil {
		return nil
	}
	if w.ResultOne == nil {
		return &hasherpb.Event_ResultOne{}
	}
	return &hasherpb.Event_ResultOne{ResultOne: (w.ResultOne).Pb()}
}

func (*Event_ResultOne) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Event_ResultOne]()}
}

func EventFromPb(pb *hasherpb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *hasherpb.Event {
	if m == nil {
		return nil
	}
	pbMessage := &hasherpb.Event{}
	{
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Event]()}
}

type Request struct {
	Data   []*HashData
	Origin *HashOrigin
}

func RequestFromPb(pb *hasherpb.Request) *Request {
	if pb == nil {
		return nil
	}
	return &Request{
		Data: types.ConvertSlice(pb.Data, func(t *hasherpb.HashData) *HashData {
			return HashDataFromPb(t)
		}),
		Origin: HashOriginFromPb(pb.Origin),
	}
}

func (m *Request) Pb() *hasherpb.Request {
	if m == nil {
		return nil
	}
	pbMessage := &hasherpb.Request{}
	{
		pbMessage.Data = types.ConvertSlice(m.Data, func(t *HashData) *hasherpb.HashData {
			return (t).Pb()
		})
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
	}

	return pbMessage
}

func (*Request) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Request]()}
}

type Result struct {
	Digests [][]uint8
	Origin  *HashOrigin
}

func ResultFromPb(pb *hasherpb.Result) *Result {
	if pb == nil {
		return nil
	}
	return &Result{
		Digests: pb.Digests,
		Origin:  HashOriginFromPb(pb.Origin),
	}
}

func (m *Result) Pb() *hasherpb.Result {
	if m == nil {
		return nil
	}
	pbMessage := &hasherpb.Result{}
	{
		pbMessage.Digests = m.Digests
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
	}

	return pbMessage
}

func (*Result) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.Result]()}
}

type RequestOne struct {
	Data   *HashData
	Origin *HashOrigin
}

func RequestOneFromPb(pb *hasherpb.RequestOne) *RequestOne {
	if pb == nil {
		return nil
	}
	return &RequestOne{
		Data:   HashDataFromPb(pb.Data),
		Origin: HashOriginFromPb(pb.Origin),
	}
}

func (m *RequestOne) Pb() *hasherpb.RequestOne {
	if m == nil {
		return nil
	}
	pbMessage := &hasherpb.RequestOne{}
	{
		if m.Data != nil {
			pbMessage.Data = (m.Data).Pb()
		}
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
	}

	return pbMessage
}

func (*RequestOne) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.RequestOne]()}
}

type ResultOne struct {
	Digest []uint8
	Origin *HashOrigin
}

func ResultOneFromPb(pb *hasherpb.ResultOne) *ResultOne {
	if pb == nil {
		return nil
	}
	return &ResultOne{
		Digest: pb.Digest,
		Origin: HashOriginFromPb(pb.Origin),
	}
}

func (m *ResultOne) Pb() *hasherpb.ResultOne {
	if m == nil {
		return nil
	}
	pbMessage := &hasherpb.ResultOne{}
	{
		pbMessage.Digest = m.Digest
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
	}

	return pbMessage
}

func (*ResultOne) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.ResultOne]()}
}

type HashOrigin struct {
	Module types1.ModuleID
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
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *hasherpb.HashOrigin_ContextStore:
		return &HashOrigin_ContextStore{ContextStore: types2.OriginFromPb(pb.ContextStore)}
	case *hasherpb.HashOrigin_Dsl:
		return &HashOrigin_Dsl{Dsl: types3.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type HashOrigin_ContextStore struct {
	ContextStore *types2.Origin
}

func (*HashOrigin_ContextStore) isHashOrigin_Type() {}

func (w *HashOrigin_ContextStore) Unwrap() *types2.Origin {
	return w.ContextStore
}

func (w *HashOrigin_ContextStore) Pb() hasherpb.HashOrigin_Type {
	if w == nil {
		return nil
	}
	if w.ContextStore == nil {
		return &hasherpb.HashOrigin_ContextStore{}
	}
	return &hasherpb.HashOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*HashOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashOrigin_ContextStore]()}
}

type HashOrigin_Dsl struct {
	Dsl *types3.Origin
}

func (*HashOrigin_Dsl) isHashOrigin_Type() {}

func (w *HashOrigin_Dsl) Unwrap() *types3.Origin {
	return w.Dsl
}

func (w *HashOrigin_Dsl) Pb() hasherpb.HashOrigin_Type {
	if w == nil {
		return nil
	}
	if w.Dsl == nil {
		return &hasherpb.HashOrigin_Dsl{}
	}
	return &hasherpb.HashOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*HashOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashOrigin_Dsl]()}
}

func HashOriginFromPb(pb *hasherpb.HashOrigin) *HashOrigin {
	if pb == nil {
		return nil
	}
	return &HashOrigin{
		Module: (types1.ModuleID)(pb.Module),
		Type:   HashOrigin_TypeFromPb(pb.Type),
	}
}

func (m *HashOrigin) Pb() *hasherpb.HashOrigin {
	if m == nil {
		return nil
	}
	pbMessage := &hasherpb.HashOrigin{}
	{
		pbMessage.Module = (string)(m.Module)
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*HashOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashOrigin]()}
}

type HashData struct {
	Data [][]uint8
}

func HashDataFromPb(pb *hasherpb.HashData) *HashData {
	if pb == nil {
		return nil
	}
	return &HashData{
		Data: pb.Data,
	}
}

func (m *HashData) Pb() *hasherpb.HashData {
	if m == nil {
		return nil
	}
	pbMessage := &hasherpb.HashData{}
	{
		pbMessage.Data = m.Data
	}

	return pbMessage
}

func (*HashData) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*hasherpb.HashData]()}
}
