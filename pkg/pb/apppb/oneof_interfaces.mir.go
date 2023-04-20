package apppb

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_SnapshotRequest) Unwrap() *SnapshotRequest {
	return w.SnapshotRequest
}

func (w *Event_Snapshot) Unwrap() *Snapshot {
	return w.Snapshot
}

func (w *Event_RestoreState) Unwrap() *RestoreState {
	return w.RestoreState
}

func (w *Event_NewEpoch) Unwrap() *NewEpoch {
	return w.NewEpoch
}
