package batchfetcherpb

import commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_NewOrderedBatch) Unwrap() *NewOrderedBatch {
	return w.NewOrderedBatch
}

func (w *Event_ClientProgress) Unwrap() *commonpb.ClientProgress {
	return w.ClientProgress
}
