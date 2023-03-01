package checkpointpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() checkpointpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb checkpointpb.Event_Type) Event_Type {
	switch pb := pb.(type) {
	case *checkpointpb.Event_EpochConfig:
		return &Event_EpochConfig{EpochConfig: pb.EpochConfig}
	case *checkpointpb.Event_StableCheckpoint:
		return &Event_StableCheckpoint{StableCheckpoint: StableCheckpointFromPb(pb.StableCheckpoint)}
	case *checkpointpb.Event_EpochProgress:
		return &Event_EpochProgress{EpochProgress: pb.EpochProgress}
	}
	return nil
}

type Event_EpochConfig struct {
	EpochConfig *commonpb.EpochConfig
}

func (*Event_EpochConfig) isEvent_Type() {}

func (w *Event_EpochConfig) Unwrap() *commonpb.EpochConfig {
	return w.EpochConfig
}

func (w *Event_EpochConfig) Pb() checkpointpb.Event_Type {
	return &checkpointpb.Event_EpochConfig{EpochConfig: w.EpochConfig}
}

func (*Event_EpochConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.Event_EpochConfig]()}
}

type Event_StableCheckpoint struct {
	StableCheckpoint *StableCheckpoint
}

func (*Event_StableCheckpoint) isEvent_Type() {}

func (w *Event_StableCheckpoint) Unwrap() *StableCheckpoint {
	return w.StableCheckpoint
}

func (w *Event_StableCheckpoint) Pb() checkpointpb.Event_Type {
	return &checkpointpb.Event_StableCheckpoint{StableCheckpoint: (w.StableCheckpoint).Pb()}
}

func (*Event_StableCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.Event_StableCheckpoint]()}
}

type Event_EpochProgress struct {
	EpochProgress *checkpointpb.EpochProgress
}

func (*Event_EpochProgress) isEvent_Type() {}

func (w *Event_EpochProgress) Unwrap() *checkpointpb.EpochProgress {
	return w.EpochProgress
}

func (w *Event_EpochProgress) Pb() checkpointpb.Event_Type {
	return &checkpointpb.Event_EpochProgress{EpochProgress: w.EpochProgress}
}

func (*Event_EpochProgress) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.Event_EpochProgress]()}
}

func EventFromPb(pb *checkpointpb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *checkpointpb.Event {
	return &checkpointpb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.Event]()}
}

type StableCheckpoint struct {
	Sn       types.SeqNr
	Snapshot *commonpb.StateSnapshot
	Cert     map[string][]uint8
}

func StableCheckpointFromPb(pb *checkpointpb.StableCheckpoint) *StableCheckpoint {
	return &StableCheckpoint{
		Sn:       (types.SeqNr)(pb.Sn),
		Snapshot: pb.Snapshot,
		Cert:     pb.Cert,
	}
}

func (m *StableCheckpoint) Pb() *checkpointpb.StableCheckpoint {
	return &checkpointpb.StableCheckpoint{
		Sn:       (uint64)(m.Sn),
		Snapshot: m.Snapshot,
		Cert:     m.Cert,
	}
}

func (*StableCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.StableCheckpoint]()}
}

type HashOrigin struct{}

func HashOriginFromPb(pb *checkpointpb.HashOrigin) *HashOrigin {
	return &HashOrigin{}
}

func (m *HashOrigin) Pb() *checkpointpb.HashOrigin {
	return &checkpointpb.HashOrigin{}
}

func (*HashOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.HashOrigin]()}
}
