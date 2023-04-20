package ordererpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	ordererpb "github.com/filecoin-project/mir/pkg/pb/ordererpb"
	types "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Message struct {
	Type Message_Type
}

type Message_Type interface {
	mirreflect.GeneratedType
	isMessage_Type()
	Pb() ordererpb.Message_Type
}

type Message_TypeWrapper[T any] interface {
	Message_Type
	Unwrap() *T
}

func Message_TypeFromPb(pb ordererpb.Message_Type) Message_Type {
	switch pb := pb.(type) {
	case *ordererpb.Message_Pbft:
		return &Message_Pbft{Pbft: types.MessageFromPb(pb.Pbft)}
	}
	return nil
}

type Message_Pbft struct {
	Pbft *types.Message
}

func (*Message_Pbft) isMessage_Type() {}

func (w *Message_Pbft) Unwrap() *types.Message {
	return w.Pbft
}

func (w *Message_Pbft) Pb() ordererpb.Message_Type {
	return &ordererpb.Message_Pbft{Pbft: (w.Pbft).Pb()}
}

func (*Message_Pbft) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*ordererpb.Message_Pbft]()}
}

func MessageFromPb(pb *ordererpb.Message) *Message {
	return &Message{
		Type: Message_TypeFromPb(pb.Type),
	}
}

func (m *Message) Pb() *ordererpb.Message {
	return &ordererpb.Message{
		Type: (m.Type).Pb(),
	}
}

func (*Message) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*ordererpb.Message]()}
}

type PBFTSegment struct {
	Leader     string
	Membership *types1.Membership
	Proposals  map[uint64][]uint8
}

func PBFTSegmentFromPb(pb *ordererpb.PBFTSegment) *PBFTSegment {
	return &PBFTSegment{
		Leader:     pb.Leader,
		Membership: types1.MembershipFromPb(pb.Membership),
		Proposals:  pb.Proposals,
	}
}

func (m *PBFTSegment) Pb() *ordererpb.PBFTSegment {
	return &ordererpb.PBFTSegment{
		Leader:     m.Leader,
		Membership: (m.Membership).Pb(),
		Proposals:  m.Proposals,
	}
}

func (*PBFTSegment) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*ordererpb.PBFTSegment]()}
}

type PBFTModule struct {
	Segment         *PBFTSegment
	AvailabilityId  string
	Epoch           uint64
	ValidityChecker uint64
}

func PBFTModuleFromPb(pb *ordererpb.PBFTModule) *PBFTModule {
	return &PBFTModule{
		Segment:         PBFTSegmentFromPb(pb.Segment),
		AvailabilityId:  pb.AvailabilityId,
		Epoch:           pb.Epoch,
		ValidityChecker: pb.ValidityChecker,
	}
}

func (m *PBFTModule) Pb() *ordererpb.PBFTModule {
	return &ordererpb.PBFTModule{
		Segment:         (m.Segment).Pb(),
		AvailabilityId:  m.AvailabilityId,
		Epoch:           m.Epoch,
		ValidityChecker: m.ValidityChecker,
	}
}

func (*PBFTModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*ordererpb.PBFTModule]()}
}
