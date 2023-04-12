package ordererspbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	ordererspb "github.com/filecoin-project/mir/pkg/pb/ordererspb"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type PBFTSegment struct {
	Leader     string
	Membership *types.Membership
	Proposals  map[uint64][]uint8
}

func PBFTSegmentFromPb(pb *ordererspb.PBFTSegment) *PBFTSegment {
	return &PBFTSegment{
		Leader:     pb.Leader,
		Membership: types.MembershipFromPb(pb.Membership),
		Proposals:  pb.Proposals,
	}
}

func (m *PBFTSegment) Pb() *ordererspb.PBFTSegment {
	return &ordererspb.PBFTSegment{
		Leader:     m.Leader,
		Membership: (m.Membership).Pb(),
		Proposals:  m.Proposals,
	}
}

func (*PBFTSegment) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*ordererspb.PBFTSegment]()}
}

type PBFTModule struct {
	Segment         *PBFTSegment
	AvailabilityId  string
	Epoch           uint64
	ValidityChecker uint64
}

func PBFTModuleFromPb(pb *ordererspb.PBFTModule) *PBFTModule {
	return &PBFTModule{
		Segment:         PBFTSegmentFromPb(pb.Segment),
		AvailabilityId:  pb.AvailabilityId,
		Epoch:           pb.Epoch,
		ValidityChecker: pb.ValidityChecker,
	}
}

func (m *PBFTModule) Pb() *ordererspb.PBFTModule {
	return &ordererspb.PBFTModule{
		Segment:         (m.Segment).Pb(),
		AvailabilityId:  m.AvailabilityId,
		Epoch:           m.Epoch,
		ValidityChecker: m.ValidityChecker,
	}
}

func (*PBFTModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*ordererspb.PBFTModule]()}
}
