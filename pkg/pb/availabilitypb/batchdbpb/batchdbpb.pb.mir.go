package batchdbpb

type Event_Type = isEvent_Type

type Event_TypeWrapper[Ev any] interface {
	Event_Type
	Unwrap() *Ev
}

func (p *Event_Lookup) Unwrap() *LookupBatch {
	return p.Lookup
}

func (p *Event_LookupResponse) Unwrap() *LookupBatchResponse {
	return p.LookupResponse
}

func (p *Event_Store) Unwrap() *StoreBatch {
	return p.Store
}

func (p *Event_Stored) Unwrap() *BatchStored {
	return p.Stored
}
