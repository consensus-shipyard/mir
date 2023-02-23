package checkpointpb

import commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_EpochConfig) Unwrap() *commonpb.EpochConfig {
	return w.EpochConfig
}

func (w *Event_StableCheckpoint) Unwrap() *StableCheckpoint {
	return w.StableCheckpoint
}

func (w *Event_EpochProgress) Unwrap() *EpochProgress {
	return w.EpochProgress
}
