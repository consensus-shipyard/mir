package hasherpb

import (
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	contextstorepb "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	dslpb "github.com/filecoin-project/mir/pkg/pb/dslpb"
	ordererpb "github.com/filecoin-project/mir/pkg/pb/ordererpb"
	trantorpb "github.com/filecoin-project/mir/pkg/pb/trantorpb"
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

func (w *HashOrigin_Request) Unwrap() *trantorpb.Transaction {
	return w.Request
}

func (w *HashOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}

func (w *HashOrigin_Checkpoint) Unwrap() *checkpointpb.HashOrigin {
	return w.Checkpoint
}

func (w *HashOrigin_Sb) Unwrap() *ordererpb.HashOrigin {
	return w.Sb
}
