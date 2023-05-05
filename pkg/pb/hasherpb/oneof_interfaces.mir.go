package hasherpb

import (
	contextstorepb "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	dslpb "github.com/filecoin-project/mir/pkg/pb/dslpb"
)

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_Request) Unwrap() *Request {
	return w.Request
}

func (w *Event_Result) Unwrap() *Result {
	return w.Result
}

func (w *Event_RequestOne) Unwrap() *RequestOne {
	return w.RequestOne
}

func (w *Event_ResultOne) Unwrap() *ResultOne {
	return w.ResultOne
}

type HashOrigin_Type = isHashOrigin_Type

type HashOrigin_TypeWrapper[T any] interface {
	HashOrigin_Type
	Unwrap() *T
}

func (w *HashOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *HashOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}
