package mempoolpb

import (
	contextstorepb "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	dslpb "github.com/filecoin-project/mir/pkg/pb/dslpb"
)

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_RequestBatch) Unwrap() *RequestBatch {
	return w.RequestBatch
}

func (w *Event_NewBatch) Unwrap() *NewBatch {
	return w.NewBatch
}

func (w *Event_RequestTransactions) Unwrap() *RequestTransactions {
	return w.RequestTransactions
}

func (w *Event_TransactionsResponse) Unwrap() *TransactionsResponse {
	return w.TransactionsResponse
}

func (w *Event_RequestTransactionIds) Unwrap() *RequestTransactionIDs {
	return w.RequestTransactionIds
}

func (w *Event_TransactionIdsResponse) Unwrap() *TransactionIDsResponse {
	return w.TransactionIdsResponse
}

func (w *Event_RequestBatchId) Unwrap() *RequestBatchID {
	return w.RequestBatchId
}

func (w *Event_BatchIdResponse) Unwrap() *BatchIDResponse {
	return w.BatchIdResponse
}

type RequestBatchOrigin_Type = isRequestBatchOrigin_Type

type RequestBatchOrigin_TypeWrapper[T any] interface {
	RequestBatchOrigin_Type
	Unwrap() *T
}

func (w *RequestBatchOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *RequestBatchOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}

type RequestTransactionsOrigin_Type = isRequestTransactionsOrigin_Type

type RequestTransactionsOrigin_TypeWrapper[T any] interface {
	RequestTransactionsOrigin_Type
	Unwrap() *T
}

func (w *RequestTransactionsOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *RequestTransactionsOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}

type RequestTransactionIDsOrigin_Type = isRequestTransactionIDsOrigin_Type

type RequestTransactionIDsOrigin_TypeWrapper[T any] interface {
	RequestTransactionIDsOrigin_Type
	Unwrap() *T
}

func (w *RequestTransactionIDsOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *RequestTransactionIDsOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}

type RequestBatchIDOrigin_Type = isRequestBatchIDOrigin_Type

type RequestBatchIDOrigin_TypeWrapper[T any] interface {
	RequestBatchIDOrigin_Type
	Unwrap() *T
}

func (w *RequestBatchIDOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *RequestBatchIDOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}
