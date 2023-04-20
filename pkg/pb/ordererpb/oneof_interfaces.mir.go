package ordererpb

import pbftpb "github.com/filecoin-project/mir/pkg/pb/pbftpb"

type Message_Type = isMessage_Type

type Message_TypeWrapper[T any] interface {
	Message_Type
	Unwrap() *T
}

func (w *Message_Pbft) Unwrap() *pbftpb.Message {
	return w.Pbft
}
