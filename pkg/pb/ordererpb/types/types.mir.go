package ordererpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types4 "github.com/filecoin-project/mir/codegen/model/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	ordererpb "github.com/filecoin-project/mir/pkg/pb/ordererpb"
	types "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	types3 "github.com/filecoin-project/mir/pkg/trantor/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
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
	Leader     types1.NodeID
	Membership *types2.Membership
	Proposals  map[types3.SeqNr][]uint8
}

func PBFTSegmentFromPb(pb *ordererpb.PBFTSegment) *PBFTSegment {
	return &PBFTSegment{
		Leader:     (types1.NodeID)(pb.Leader),
		Membership: types2.MembershipFromPb(pb.Membership),
		Proposals: types4.ConvertMap(pb.Proposals, func(k uint64, v []uint8) (types3.SeqNr, []uint8) {
			return (types3.SeqNr)(k), v
		}),
	}
}

func (m *PBFTSegment) Pb() *ordererpb.PBFTSegment {
	return &ordererpb.PBFTSegment{
		Leader:     (string)(m.Leader),
		Membership: (m.Membership).Pb(),
		Proposals: types4.ConvertMap(m.Proposals, func(k types3.SeqNr, v []uint8) (uint64, []uint8) {
			return (uint64)(k), v
		}),
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
