package commonpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types "github.com/filecoin-project/mir/codegen/model/types"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	types1 "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type HashData struct {
	Data [][]uint8
}

func HashDataFromPb(pb *commonpb.HashData) *HashData {
	return &HashData{
		Data: pb.Data,
	}
}

func (m *HashData) Pb() *commonpb.HashData {
	return &commonpb.HashData{
		Data: m.Data,
	}
}

func (*HashData) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.HashData]()}
}

type EpochData struct {
	EpochConfig        *EpochConfig
	ClientProgress     *ClientProgress
	LeaderPolicy       []uint8
	PreviousMembership *Membership
}

func EpochDataFromPb(pb *commonpb.EpochData) *EpochData {
	return &EpochData{
		EpochConfig:        EpochConfigFromPb(pb.EpochConfig),
		ClientProgress:     ClientProgressFromPb(pb.ClientProgress),
		LeaderPolicy:       pb.LeaderPolicy,
		PreviousMembership: MembershipFromPb(pb.PreviousMembership),
	}
}

func (m *EpochData) Pb() *commonpb.EpochData {
	return &commonpb.EpochData{
		EpochConfig:        (m.EpochConfig).Pb(),
		ClientProgress:     (m.ClientProgress).Pb(),
		LeaderPolicy:       m.LeaderPolicy,
		PreviousMembership: (m.PreviousMembership).Pb(),
	}
}

func (*EpochData) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.EpochData]()}
}

type EpochConfig struct {
	EpochNr     uint64
	FirstSn     uint64
	Length      uint64
	Memberships []*Membership
}

func EpochConfigFromPb(pb *commonpb.EpochConfig) *EpochConfig {
	return &EpochConfig{
		EpochNr: pb.EpochNr,
		FirstSn: pb.FirstSn,
		Length:  pb.Length,
		Memberships: types.ConvertSlice(pb.Memberships, func(t *commonpb.Membership) *Membership {
			return MembershipFromPb(t)
		}),
	}
}

func (m *EpochConfig) Pb() *commonpb.EpochConfig {
	return &commonpb.EpochConfig{
		EpochNr: m.EpochNr,
		FirstSn: m.FirstSn,
		Length:  m.Length,
		Memberships: types.ConvertSlice(m.Memberships, func(t *Membership) *commonpb.Membership {
			return (t).Pb()
		}),
	}
}

func (*EpochConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.EpochConfig]()}
}

type Membership struct {
	Membership map[types1.NodeID]string
}

func MembershipFromPb(pb *commonpb.Membership) *Membership {
	return &Membership{
		Membership: types.ConvertMap(pb.Membership, func(k string, v string) (types1.NodeID, string) {
			return (types1.NodeID)(k), v
		}),
	}
}

func (m *Membership) Pb() *commonpb.Membership {
	return &commonpb.Membership{
		Membership: types.ConvertMap(m.Membership, func(k types1.NodeID, v string) (string, string) {
			return (string)(k), v
		}),
	}
}

func (*Membership) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.Membership]()}
}

type ClientProgress struct {
	Progress map[string]*commonpb.DeliveredReqs
}

func ClientProgressFromPb(pb *commonpb.ClientProgress) *ClientProgress {
	return &ClientProgress{
		Progress: pb.Progress,
	}
}

func (m *ClientProgress) Pb() *commonpb.ClientProgress {
	return &commonpb.ClientProgress{
		Progress: m.Progress,
	}
}

func (*ClientProgress) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.ClientProgress]()}
}
