package batchdbpb

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_Lookup) Unwrap() *LookupBatch {
	return w.Lookup
}

func (w *Event_LookupResponse) Unwrap() *LookupBatchResponse {
	return w.LookupResponse
}

func (w *Event_Store) Unwrap() *StoreBatch {
	return w.Store
}

func (w *Event_Stored) Unwrap() *BatchStored {
	return w.Stored
}
