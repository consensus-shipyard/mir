package transportpb

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_SendMessage) Unwrap() *SendMessage {
	return w.SendMessage
}

func (w *Event_MessageReceived) Unwrap() *MessageReceived {
	return w.MessageReceived
}
