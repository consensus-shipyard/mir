package isspbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types3 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	types4 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type ISSMessage struct {
	Type ISSMessage_Type
}

type ISSMessage_Type interface {
	mirreflect.GeneratedType
	isISSMessage_Type()
	Pb() isspb.ISSMessage_Type
}

type ISSMessage_TypeWrapper[T any] interface {
	ISSMessage_Type
	Unwrap() *T
}

func ISSMessage_TypeFromPb(pb isspb.ISSMessage_Type) ISSMessage_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *isspb.ISSMessage_StableCheckpoint:
		return &ISSMessage_StableCheckpoint{StableCheckpoint: types.StableCheckpointFromPb(pb.StableCheckpoint)}
	}
	return nil
}

type ISSMessage_StableCheckpoint struct {
	StableCheckpoint *types.StableCheckpoint
}

func (*ISSMessage_StableCheckpoint) isISSMessage_Type() {}

func (w *ISSMessage_StableCheckpoint) Unwrap() *types.StableCheckpoint {
	return w.StableCheckpoint
}

func (w *ISSMessage_StableCheckpoint) Pb() isspb.ISSMessage_Type {
	if w == nil {
		return nil
	}
	return &isspb.ISSMessage_StableCheckpoint{StableCheckpoint: (w.StableCheckpoint).Pb()}
}

func (*ISSMessage_StableCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSMessage_StableCheckpoint]()}
}

func ISSMessageFromPb(pb *isspb.ISSMessage) *ISSMessage {
	if pb == nil {
		return nil
	}
	return &ISSMessage{
		Type: ISSMessage_TypeFromPb(pb.Type),
	}
}

func (m *ISSMessage) Pb() *isspb.ISSMessage {
	if m == nil {
		return nil
	}
	return &isspb.ISSMessage{
		Type: (m.Type).Pb(),
	}
}

func (*ISSMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSMessage]()}
}

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() isspb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb isspb.Event_Type) Event_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *isspb.Event_PushCheckpoint:
		return &Event_PushCheckpoint{PushCheckpoint: PushCheckpointFromPb(pb.PushCheckpoint)}
	case *isspb.Event_SbDeliver:
		return &Event_SbDeliver{SbDeliver: SBDeliverFromPb(pb.SbDeliver)}
	case *isspb.Event_DeliverCert:
		return &Event_DeliverCert{DeliverCert: DeliverCertFromPb(pb.DeliverCert)}
	case *isspb.Event_NewConfig:
		return &Event_NewConfig{NewConfig: NewConfigFromPb(pb.NewConfig)}
	}
	return nil
}

type Event_PushCheckpoint struct {
	PushCheckpoint *PushCheckpoint
}

func (*Event_PushCheckpoint) isEvent_Type() {}

func (w *Event_PushCheckpoint) Unwrap() *PushCheckpoint {
	return w.PushCheckpoint
}

func (w *Event_PushCheckpoint) Pb() isspb.Event_Type {
	if w == nil {
		return nil
	}
	return &isspb.Event_PushCheckpoint{PushCheckpoint: (w.PushCheckpoint).Pb()}
}

func (*Event_PushCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.Event_PushCheckpoint]()}
}

type Event_SbDeliver struct {
	SbDeliver *SBDeliver
}

func (*Event_SbDeliver) isEvent_Type() {}

func (w *Event_SbDeliver) Unwrap() *SBDeliver {
	return w.SbDeliver
}

func (w *Event_SbDeliver) Pb() isspb.Event_Type {
	if w == nil {
		return nil
	}
	return &isspb.Event_SbDeliver{SbDeliver: (w.SbDeliver).Pb()}
}

func (*Event_SbDeliver) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.Event_SbDeliver]()}
}

type Event_DeliverCert struct {
	DeliverCert *DeliverCert
}

func (*Event_DeliverCert) isEvent_Type() {}

func (w *Event_DeliverCert) Unwrap() *DeliverCert {
	return w.DeliverCert
}

func (w *Event_DeliverCert) Pb() isspb.Event_Type {
	if w == nil {
		return nil
	}
	return &isspb.Event_DeliverCert{DeliverCert: (w.DeliverCert).Pb()}
}

func (*Event_DeliverCert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.Event_DeliverCert]()}
}

type Event_NewConfig struct {
	NewConfig *NewConfig
}

func (*Event_NewConfig) isEvent_Type() {}

func (w *Event_NewConfig) Unwrap() *NewConfig {
	return w.NewConfig
}

func (w *Event_NewConfig) Pb() isspb.Event_Type {
	if w == nil {
		return nil
	}
	return &isspb.Event_NewConfig{NewConfig: (w.NewConfig).Pb()}
}

func (*Event_NewConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.Event_NewConfig]()}
}

func EventFromPb(pb *isspb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *isspb.Event {
	if m == nil {
		return nil
	}
	return &isspb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.Event]()}
}

type PushCheckpoint struct{}

func PushCheckpointFromPb(pb *isspb.PushCheckpoint) *PushCheckpoint {
	if pb == nil {
		return nil
	}
	return &PushCheckpoint{}
}

func (m *PushCheckpoint) Pb() *isspb.PushCheckpoint {
	if m == nil {
		return nil
	}
	return &isspb.PushCheckpoint{}
}

func (*PushCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.PushCheckpoint]()}
}

type SBDeliver struct {
	Sn         types1.SeqNr
	Data       []uint8
	Aborted    bool
	Leader     types2.NodeID
	InstanceId types2.ModuleID
}

func SBDeliverFromPb(pb *isspb.SBDeliver) *SBDeliver {
	if pb == nil {
		return nil
	}
	return &SBDeliver{
		Sn:         (types1.SeqNr)(pb.Sn),
		Data:       pb.Data,
		Aborted:    pb.Aborted,
		Leader:     (types2.NodeID)(pb.Leader),
		InstanceId: (types2.ModuleID)(pb.InstanceId),
	}
}

func (m *SBDeliver) Pb() *isspb.SBDeliver {
	if m == nil {
		return nil
	}
	return &isspb.SBDeliver{
		Sn:         (uint64)(m.Sn),
		Data:       m.Data,
		Aborted:    m.Aborted,
		Leader:     (string)(m.Leader),
		InstanceId: (string)(m.InstanceId),
	}
}

func (*SBDeliver) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.SBDeliver]()}
}

type DeliverCert struct {
	Sn    types1.SeqNr
	Cert  *types3.Cert
	Empty bool
}

func DeliverCertFromPb(pb *isspb.DeliverCert) *DeliverCert {
	if pb == nil {
		return nil
	}
	return &DeliverCert{
		Sn:    (types1.SeqNr)(pb.Sn),
		Cert:  types3.CertFromPb(pb.Cert),
		Empty: pb.Empty,
	}
}

func (m *DeliverCert) Pb() *isspb.DeliverCert {
	if m == nil {
		return nil
	}
	return &isspb.DeliverCert{
		Sn:    (uint64)(m.Sn),
		Cert:  (m.Cert).Pb(),
		Empty: m.Empty,
	}
}

func (*DeliverCert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.DeliverCert]()}
}

type NewConfig struct {
	EpochNr    types1.EpochNr
	Membership *types4.Membership
}

func NewConfigFromPb(pb *isspb.NewConfig) *NewConfig {
	if pb == nil {
		return nil
	}
	return &NewConfig{
		EpochNr:    (types1.EpochNr)(pb.EpochNr),
		Membership: types4.MembershipFromPb(pb.Membership),
	}
}

func (m *NewConfig) Pb() *isspb.NewConfig {
	if m == nil {
		return nil
	}
	return &isspb.NewConfig{
		EpochNr:    (uint64)(m.EpochNr),
		Membership: (m.Membership).Pb(),
	}
}

func (*NewConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.NewConfig]()}
}
