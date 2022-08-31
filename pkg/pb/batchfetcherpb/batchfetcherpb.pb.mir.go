package batchfetcherpb

type Event_Type = isEvent_Type

type Event_TypeWrapper[Ev any] interface {
	Event_Type
	Unwrap() *Ev
}

func (p *Event_NewOrderedBatch) Unwrap() *NewOrderedBatch {
	return p.NewOrderedBatch
}
