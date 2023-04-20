package isspb

import checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"

type ISSMessage_Type = isISSMessage_Type

type ISSMessage_TypeWrapper[T any] interface {
	ISSMessage_Type
	Unwrap() *T
}

func (w *ISSMessage_StableCheckpoint) Unwrap() *checkpointpb.StableCheckpoint {
	return w.StableCheckpoint
}

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_PushCheckpoint) Unwrap() *PushCheckpoint {
	return w.PushCheckpoint
}

func (w *Event_SbDeliver) Unwrap() *SBDeliver {
	return w.SbDeliver
}

func (w *Event_DeliverCert) Unwrap() *DeliverCert {
	return w.DeliverCert
}

func (w *Event_NewConfig) Unwrap() *NewConfig {
	return w.NewConfig
}
