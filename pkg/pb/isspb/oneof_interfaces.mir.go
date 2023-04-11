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

func (w *ISSMessage_RetransmitRequests) Unwrap() *RetransmitRequests {
	return w.RetransmitRequests
}

type ISSEvent_Type = isISSEvent_Type

type ISSEvent_TypeWrapper[T any] interface {
	ISSEvent_Type
	Unwrap() *T
}

func (w *ISSEvent_PersistCheckpoint) Unwrap() *PersistCheckpoint {
	return w.PersistCheckpoint
}

func (w *ISSEvent_PersistStableCheckpoint) Unwrap() *PersistStableCheckpoint {
	return w.PersistStableCheckpoint
}

func (w *ISSEvent_PushCheckpoint) Unwrap() *PushCheckpoint {
	return w.PushCheckpoint
}

func (w *ISSEvent_SbDeliver) Unwrap() *SBDeliver {
	return w.SbDeliver
}
