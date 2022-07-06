package bcbpb

type Event_Type = isEvent_Type

type Event_TypeWrapper[Ev any] interface {
	Event_Type
	Unwrap() *Ev
}

func (p *Event_Request) Unwrap() *Request {
	return p.Request
}

func (p *Event_Deliver) Unwrap() *Deliver {
	return p.Deliver
}
