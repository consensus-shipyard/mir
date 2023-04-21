package checkpointpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types3 "github.com/filecoin-project/mir/codegen/model/types"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types4 "github.com/filecoin-project/mir/pkg/timer/types"
	types "github.com/filecoin-project/mir/pkg/trantor/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Message struct {
	Type Message_Type
}

type Message_Type interface {
	mirreflect.GeneratedType
	isMessage_Type()
	Pb() checkpointpb.Message_Type
}

type Message_TypeWrapper[T any] interface {
	Message_Type
	Unwrap() *T
}

func Message_TypeFromPb(pb checkpointpb.Message_Type) Message_Type {
	switch pb := pb.(type) {
	case *checkpointpb.Message_Checkpoint:
		return &Message_Checkpoint{Checkpoint: CheckpointFromPb(pb.Checkpoint)}
	}
	return nil
}

type Message_Checkpoint struct {
	Checkpoint *Checkpoint
}

func (*Message_Checkpoint) isMessage_Type() {}

func (w *Message_Checkpoint) Unwrap() *Checkpoint {
	return w.Checkpoint
}

func (w *Message_Checkpoint) Pb() checkpointpb.Message_Type {
	return &checkpointpb.Message_Checkpoint{Checkpoint: (w.Checkpoint).Pb()}
}

func (*Message_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.Message_Checkpoint]()}
}

func MessageFromPb(pb *checkpointpb.Message) *Message {
	return &Message{
		Type: Message_TypeFromPb(pb.Type),
	}
}

func (m *Message) Pb() *checkpointpb.Message {
	return &checkpointpb.Message{
		Type: (m.Type).Pb(),
	}
}

func (*Message) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.Message]()}
}

type Checkpoint struct {
	Epoch        types.EpochNr
	Sn           types.SeqNr
	SnapshotHash []uint8
	Signature    []uint8
}

func CheckpointFromPb(pb *checkpointpb.Checkpoint) *Checkpoint {
	return &Checkpoint{
		Epoch:        (types.EpochNr)(pb.Epoch),
		Sn:           (types.SeqNr)(pb.Sn),
		SnapshotHash: pb.SnapshotHash,
		Signature:    pb.Signature,
	}
}

func (m *Checkpoint) Pb() *checkpointpb.Checkpoint {
	return &checkpointpb.Checkpoint{
		Epoch:        (uint64)(m.Epoch),
		Sn:           (uint64)(m.Sn),
		SnapshotHash: m.SnapshotHash,
		Signature:    m.Signature,
	}
}

func (*Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.Checkpoint]()}
}

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
		return &Event_EpochConfig{EpochConfig: types1.EpochConfigFromPb(pb.EpochConfig)}
	case *checkpointpb.Event_StableCheckpoint:
		return &Event_StableCheckpoint{StableCheckpoint: StableCheckpointFromPb(pb.StableCheckpoint)}
	case *checkpointpb.Event_EpochProgress:
		return &Event_EpochProgress{EpochProgress: EpochProgressFromPb(pb.EpochProgress)}
	}
	return nil
}

type Event_EpochConfig struct {
	EpochConfig *types1.EpochConfig
}

func (*Event_EpochConfig) isEvent_Type() {}

func (w *Event_EpochConfig) Unwrap() *types1.EpochConfig {
	return w.EpochConfig
}

func (w *Event_EpochConfig) Pb() checkpointpb.Event_Type {
	return &checkpointpb.Event_EpochConfig{EpochConfig: (w.EpochConfig).Pb()}
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
	EpochProgress *EpochProgress
}

func (*Event_EpochProgress) isEvent_Type() {}

func (w *Event_EpochProgress) Unwrap() *EpochProgress {
	return w.EpochProgress
}

func (w *Event_EpochProgress) Pb() checkpointpb.Event_Type {
	return &checkpointpb.Event_EpochProgress{EpochProgress: (w.EpochProgress).Pb()}
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
	Snapshot *types1.StateSnapshot
	Cert     map[types2.NodeID][]uint8
}

func StableCheckpointFromPb(pb *checkpointpb.StableCheckpoint) *StableCheckpoint {
	return &StableCheckpoint{
		Sn:       (types.SeqNr)(pb.Sn),
		Snapshot: types1.StateSnapshotFromPb(pb.Snapshot),
		Cert: types3.ConvertMap(pb.Cert, func(k string, v []uint8) (types2.NodeID, []uint8) {
			return (types2.NodeID)(k), v
		}),
	}
}

func (m *StableCheckpoint) Pb() *checkpointpb.StableCheckpoint {
	return &checkpointpb.StableCheckpoint{
		Sn:       (uint64)(m.Sn),
		Snapshot: (m.Snapshot).Pb(),
		Cert: types3.ConvertMap(m.Cert, func(k types2.NodeID, v []uint8) (string, []uint8) {
			return (string)(k), v
		}),
	}
}

func (*StableCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.StableCheckpoint]()}
}

type EpochProgress struct {
	NodeId types2.NodeID
	Epoch  types.EpochNr
}

func EpochProgressFromPb(pb *checkpointpb.EpochProgress) *EpochProgress {
	return &EpochProgress{
		NodeId: (types2.NodeID)(pb.NodeId),
		Epoch:  (types.EpochNr)(pb.Epoch),
	}
}

func (m *EpochProgress) Pb() *checkpointpb.EpochProgress {
	return &checkpointpb.EpochProgress{
		NodeId: (string)(m.NodeId),
		Epoch:  (uint64)(m.Epoch),
	}
}

func (*EpochProgress) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.EpochProgress]()}
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

type SignOrigin struct{}

func SignOriginFromPb(pb *checkpointpb.SignOrigin) *SignOrigin {
	return &SignOrigin{}
}

func (m *SignOrigin) Pb() *checkpointpb.SignOrigin {
	return &checkpointpb.SignOrigin{}
}

func (*SignOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.SignOrigin]()}
}

type SigVerOrigin struct{}

func SigVerOriginFromPb(pb *checkpointpb.SigVerOrigin) *SigVerOrigin {
	return &SigVerOrigin{}
}

func (m *SigVerOrigin) Pb() *checkpointpb.SigVerOrigin {
	return &checkpointpb.SigVerOrigin{}
}

func (*SigVerOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.SigVerOrigin]()}
}

type InstanceParams struct {
	Membership       *types1.Membership
	ResendPeriod     types4.Duration
	LeaderPolicyData []uint8
	EpochConfig      *types1.EpochConfig
}

func InstanceParamsFromPb(pb *checkpointpb.InstanceParams) *InstanceParams {
	return &InstanceParams{
		Membership:       types1.MembershipFromPb(pb.Membership),
		ResendPeriod:     (types4.Duration)(pb.ResendPeriod),
		LeaderPolicyData: pb.LeaderPolicyData,
		EpochConfig:      types1.EpochConfigFromPb(pb.EpochConfig),
	}
}

func (m *InstanceParams) Pb() *checkpointpb.InstanceParams {
	return &checkpointpb.InstanceParams{
		Membership:       (m.Membership).Pb(),
		ResendPeriod:     (uint64)(m.ResendPeriod),
		LeaderPolicyData: m.LeaderPolicyData,
		EpochConfig:      (m.EpochConfig).Pb(),
	}
}

func (*InstanceParams) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.InstanceParams]()}
}
