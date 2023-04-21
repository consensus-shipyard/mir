package isspbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types2 "github.com/filecoin-project/mir/codegen/model/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types1 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types3 "github.com/filecoin-project/mir/pkg/trantor/types"
	types4 "github.com/filecoin-project/mir/pkg/types"
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
	return &isspb.ISSMessage_StableCheckpoint{StableCheckpoint: (w.StableCheckpoint).Pb()}
}

func (*ISSMessage_StableCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSMessage_StableCheckpoint]()}
}

func ISSMessageFromPb(pb *isspb.ISSMessage) *ISSMessage {
	return &ISSMessage{
		Type: ISSMessage_TypeFromPb(pb.Type),
	}
}

func (m *ISSMessage) Pb() *isspb.ISSMessage {
	return &isspb.ISSMessage{
		Type: (m.Type).Pb(),
	}
}

func (*ISSMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSMessage]()}
}

type RetransmitRequests struct {
	Requests []*types1.Request
}

func RetransmitRequestsFromPb(pb *isspb.RetransmitRequests) *RetransmitRequests {
	return &RetransmitRequests{
		Requests: types2.ConvertSlice(pb.Requests, func(t *requestpb.Request) *types1.Request {
			return types1.RequestFromPb(t)
		}),
	}
}

func (m *RetransmitRequests) Pb() *isspb.RetransmitRequests {
	return &isspb.RetransmitRequests{
		Requests: types2.ConvertSlice(m.Requests, func(t *types1.Request) *requestpb.Request {
			return (t).Pb()
		}),
	}
}

func (*RetransmitRequests) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.RetransmitRequests]()}
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
	return &isspb.Event_NewConfig{NewConfig: (w.NewConfig).Pb()}
}

func (*Event_NewConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.Event_NewConfig]()}
}

func EventFromPb(pb *isspb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *isspb.Event {
	return &isspb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.Event]()}
}

type PushCheckpoint struct{}

func PushCheckpointFromPb(pb *isspb.PushCheckpoint) *PushCheckpoint {
	return &PushCheckpoint{}
}

func (m *PushCheckpoint) Pb() *isspb.PushCheckpoint {
	return &isspb.PushCheckpoint{}
}

func (*PushCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.PushCheckpoint]()}
}

type SBDeliver struct {
	Sn         types3.SeqNr
	Data       []uint8
	Aborted    bool
	Leader     types4.NodeID
	InstanceId types4.ModuleID
}

func SBDeliverFromPb(pb *isspb.SBDeliver) *SBDeliver {
	return &SBDeliver{
		Sn:         (types3.SeqNr)(pb.Sn),
		Data:       pb.Data,
		Aborted:    pb.Aborted,
		Leader:     (types4.NodeID)(pb.Leader),
		InstanceId: (types4.ModuleID)(pb.InstanceId),
	}
}

func (m *SBDeliver) Pb() *isspb.SBDeliver {
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
	Sn   types3.SeqNr
	Cert *types5.Cert
}

func DeliverCertFromPb(pb *isspb.DeliverCert) *DeliverCert {
	return &DeliverCert{
		Sn:   (types3.SeqNr)(pb.Sn),
		Cert: types5.CertFromPb(pb.Cert),
	}
}

func (m *DeliverCert) Pb() *isspb.DeliverCert {
	return &isspb.DeliverCert{
		Sn:   (uint64)(m.Sn),
		Cert: (m.Cert).Pb(),
	}
}

func (*DeliverCert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.DeliverCert]()}
}

type NewConfig struct {
	EpochNr    types3.EpochNr
	Membership *types6.Membership
}

func NewConfigFromPb(pb *isspb.NewConfig) *NewConfig {
	return &NewConfig{
		EpochNr:    (types3.EpochNr)(pb.EpochNr),
		Membership: types6.MembershipFromPb(pb.Membership),
	}
}

func (m *NewConfig) Pb() *isspb.NewConfig {
	return &isspb.NewConfig{
		EpochNr:    (uint64)(m.EpochNr),
		Membership: (m.Membership).Pb(),
	}
}

func (*NewConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.NewConfig]()}
}
