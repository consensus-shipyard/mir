package ordererpb

import pbftpb "github.com/filecoin-project/mir/pkg/pb/pbftpb"

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_Pbft) Unwrap() *pbftpb.Event {
	return w.Pbft
}

type Message_Type = isMessage_Type

type Message_TypeWrapper[T any] interface {
	Message_Type
	Unwrap() *T
}

func (w *Message_Pbft) Unwrap() *pbftpb.Message {
	return w.Pbft
}

type SignOrigin_Type = isSignOrigin_Type

type SignOrigin_TypeWrapper[T any] interface {
	SignOrigin_Type
	Unwrap() *T
}

func (w *SignOrigin_Pbft) Unwrap() *pbftpb.SignOrigin {
	return w.Pbft
}
