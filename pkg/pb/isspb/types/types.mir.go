package isspbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types2 "github.com/filecoin-project/mir/codegen/model/types"
	types "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types1 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types3 "github.com/filecoin-project/mir/pkg/types"
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
	case *isspb.ISSMessage_RetransmitRequests:
		return &ISSMessage_RetransmitRequests{RetransmitRequests: RetransmitRequestsFromPb(pb.RetransmitRequests)}
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

type ISSMessage_RetransmitRequests struct {
	RetransmitRequests *RetransmitRequests
}

func (*ISSMessage_RetransmitRequests) isISSMessage_Type() {}

func (w *ISSMessage_RetransmitRequests) Unwrap() *RetransmitRequests {
	return w.RetransmitRequests
}

func (w *ISSMessage_RetransmitRequests) Pb() isspb.ISSMessage_Type {
	return &isspb.ISSMessage_RetransmitRequests{RetransmitRequests: (w.RetransmitRequests).Pb()}
}

func (*ISSMessage_RetransmitRequests) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSMessage_RetransmitRequests]()}
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

type ISSEvent struct {
	Type ISSEvent_Type
}

type ISSEvent_Type interface {
	mirreflect.GeneratedType
	isISSEvent_Type()
	Pb() isspb.ISSEvent_Type
}

type ISSEvent_TypeWrapper[T any] interface {
	ISSEvent_Type
	Unwrap() *T
}

func ISSEvent_TypeFromPb(pb isspb.ISSEvent_Type) ISSEvent_Type {
	switch pb := pb.(type) {
	case *isspb.ISSEvent_PersistCheckpoint:
		return &ISSEvent_PersistCheckpoint{PersistCheckpoint: pb.PersistCheckpoint}
	case *isspb.ISSEvent_PersistStableCheckpoint:
		return &ISSEvent_PersistStableCheckpoint{PersistStableCheckpoint: pb.PersistStableCheckpoint}
	case *isspb.ISSEvent_PushCheckpoint:
		return &ISSEvent_PushCheckpoint{PushCheckpoint: PushCheckpointFromPb(pb.PushCheckpoint)}
	case *isspb.ISSEvent_SbDeliver:
		return &ISSEvent_SbDeliver{SbDeliver: SBDeliverFromPb(pb.SbDeliver)}
	}
	return nil
}

type ISSEvent_PersistCheckpoint struct {
	PersistCheckpoint *isspb.PersistCheckpoint
}

func (*ISSEvent_PersistCheckpoint) isISSEvent_Type() {}

func (w *ISSEvent_PersistCheckpoint) Unwrap() *isspb.PersistCheckpoint {
	return w.PersistCheckpoint
}

func (w *ISSEvent_PersistCheckpoint) Pb() isspb.ISSEvent_Type {
	return &isspb.ISSEvent_PersistCheckpoint{PersistCheckpoint: w.PersistCheckpoint}
}

func (*ISSEvent_PersistCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSEvent_PersistCheckpoint]()}
}

type ISSEvent_PersistStableCheckpoint struct {
	PersistStableCheckpoint *isspb.PersistStableCheckpoint
}

func (*ISSEvent_PersistStableCheckpoint) isISSEvent_Type() {}

func (w *ISSEvent_PersistStableCheckpoint) Unwrap() *isspb.PersistStableCheckpoint {
	return w.PersistStableCheckpoint
}

func (w *ISSEvent_PersistStableCheckpoint) Pb() isspb.ISSEvent_Type {
	return &isspb.ISSEvent_PersistStableCheckpoint{PersistStableCheckpoint: w.PersistStableCheckpoint}
}

func (*ISSEvent_PersistStableCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSEvent_PersistStableCheckpoint]()}
}

type ISSEvent_PushCheckpoint struct {
	PushCheckpoint *PushCheckpoint
}

func (*ISSEvent_PushCheckpoint) isISSEvent_Type() {}

func (w *ISSEvent_PushCheckpoint) Unwrap() *PushCheckpoint {
	return w.PushCheckpoint
}

func (w *ISSEvent_PushCheckpoint) Pb() isspb.ISSEvent_Type {
	return &isspb.ISSEvent_PushCheckpoint{PushCheckpoint: (w.PushCheckpoint).Pb()}
}

func (*ISSEvent_PushCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSEvent_PushCheckpoint]()}
}

type ISSEvent_SbDeliver struct {
	SbDeliver *SBDeliver
}

func (*ISSEvent_SbDeliver) isISSEvent_Type() {}

func (w *ISSEvent_SbDeliver) Unwrap() *SBDeliver {
	return w.SbDeliver
}

func (w *ISSEvent_SbDeliver) Pb() isspb.ISSEvent_Type {
	return &isspb.ISSEvent_SbDeliver{SbDeliver: (w.SbDeliver).Pb()}
}

func (*ISSEvent_SbDeliver) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSEvent_SbDeliver]()}
}

func ISSEventFromPb(pb *isspb.ISSEvent) *ISSEvent {
	return &ISSEvent{
		Type: ISSEvent_TypeFromPb(pb.Type),
	}
}

func (m *ISSEvent) Pb() *isspb.ISSEvent {
	return &isspb.ISSEvent{
		Type: (m.Type).Pb(),
	}
}

func (*ISSEvent) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*isspb.ISSEvent]()}
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
	Leader     types3.NodeID
	InstanceId types3.ModuleID
}

func SBDeliverFromPb(pb *isspb.SBDeliver) *SBDeliver {
	return &SBDeliver{
		Sn:         (types3.SeqNr)(pb.Sn),
		Data:       pb.Data,
		Aborted:    pb.Aborted,
		Leader:     (types3.NodeID)(pb.Leader),
		InstanceId: (types3.ModuleID)(pb.InstanceId),
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
