package apppbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	apppb "github.com/filecoin-project/mir/pkg/pb/apppb"
	types1 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types2 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() apppb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb apppb.Event_Type) Event_Type {
	switch pb := pb.(type) {
	case *apppb.Event_SnapshotRequest:
		return &Event_SnapshotRequest{SnapshotRequest: SnapshotRequestFromPb(pb.SnapshotRequest)}
	case *apppb.Event_Snapshot:
		return &Event_Snapshot{Snapshot: SnapshotFromPb(pb.Snapshot)}
	case *apppb.Event_RestoreState:
		return &Event_RestoreState{RestoreState: RestoreStateFromPb(pb.RestoreState)}
	case *apppb.Event_NewEpoch:
		return &Event_NewEpoch{NewEpoch: NewEpochFromPb(pb.NewEpoch)}
	}
	return nil
}

type Event_SnapshotRequest struct {
	SnapshotRequest *SnapshotRequest
}

func (*Event_SnapshotRequest) isEvent_Type() {}

func (w *Event_SnapshotRequest) Unwrap() *SnapshotRequest {
	return w.SnapshotRequest
}

func (w *Event_SnapshotRequest) Pb() apppb.Event_Type {
	return &apppb.Event_SnapshotRequest{SnapshotRequest: (w.SnapshotRequest).Pb()}
}

func (*Event_SnapshotRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*apppb.Event_SnapshotRequest]()}
}

type Event_Snapshot struct {
	Snapshot *Snapshot
}

func (*Event_Snapshot) isEvent_Type() {}

func (w *Event_Snapshot) Unwrap() *Snapshot {
	return w.Snapshot
}

func (w *Event_Snapshot) Pb() apppb.Event_Type {
	return &apppb.Event_Snapshot{Snapshot: (w.Snapshot).Pb()}
}

func (*Event_Snapshot) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*apppb.Event_Snapshot]()}
}

type Event_RestoreState struct {
	RestoreState *RestoreState
}

func (*Event_RestoreState) isEvent_Type() {}

func (w *Event_RestoreState) Unwrap() *RestoreState {
	return w.RestoreState
}

func (w *Event_RestoreState) Pb() apppb.Event_Type {
	return &apppb.Event_RestoreState{RestoreState: (w.RestoreState).Pb()}
}

func (*Event_RestoreState) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*apppb.Event_RestoreState]()}
}

type Event_NewEpoch struct {
	NewEpoch *NewEpoch
}

func (*Event_NewEpoch) isEvent_Type() {}

func (w *Event_NewEpoch) Unwrap() *NewEpoch {
	return w.NewEpoch
}

func (w *Event_NewEpoch) Pb() apppb.Event_Type {
	return &apppb.Event_NewEpoch{NewEpoch: (w.NewEpoch).Pb()}
}

func (*Event_NewEpoch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*apppb.Event_NewEpoch]()}
}

func EventFromPb(pb *apppb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *apppb.Event {
	return &apppb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*apppb.Event]()}
}

type SnapshotRequest struct {
	ReplyTo types.ModuleID
}

func SnapshotRequestFromPb(pb *apppb.SnapshotRequest) *SnapshotRequest {
	return &SnapshotRequest{
		ReplyTo: (types.ModuleID)(pb.ReplyTo),
	}
}

func (m *SnapshotRequest) Pb() *apppb.SnapshotRequest {
	return &apppb.SnapshotRequest{
		ReplyTo: (string)(m.ReplyTo),
	}
}

func (*SnapshotRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*apppb.SnapshotRequest]()}
}

type Snapshot struct {
	AppData []uint8
}

func SnapshotFromPb(pb *apppb.Snapshot) *Snapshot {
	return &Snapshot{
		AppData: pb.AppData,
	}
}

func (m *Snapshot) Pb() *apppb.Snapshot {
	return &apppb.Snapshot{
		AppData: m.AppData,
	}
}

func (*Snapshot) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*apppb.Snapshot]()}
}

type RestoreState struct {
	Checkpoint *types1.StableCheckpoint
}

func RestoreStateFromPb(pb *apppb.RestoreState) *RestoreState {
	return &RestoreState{
		Checkpoint: types1.StableCheckpointFromPb(pb.Checkpoint),
	}
}

func (m *RestoreState) Pb() *apppb.RestoreState {
	return &apppb.RestoreState{
		Checkpoint: (m.Checkpoint).Pb(),
	}
}

func (*RestoreState) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*apppb.RestoreState]()}
}

type NewEpoch struct {
	EpochNr types2.EpochNr
}

func NewEpochFromPb(pb *apppb.NewEpoch) *NewEpoch {
	return &NewEpoch{
		EpochNr: (types2.EpochNr)(pb.EpochNr),
	}
}

func (m *NewEpoch) Pb() *apppb.NewEpoch {
	return &apppb.NewEpoch{
		EpochNr: (uint64)(m.EpochNr),
	}
}

func (*NewEpoch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*apppb.NewEpoch]()}
}
