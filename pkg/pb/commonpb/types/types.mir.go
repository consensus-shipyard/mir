package commonpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	types "github.com/filecoin-project/mir/pkg/trantor/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
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

type StateSnapshot struct {
	AppData   []uint8
	EpochData *EpochData
}

func StateSnapshotFromPb(pb *commonpb.StateSnapshot) *StateSnapshot {
	return &StateSnapshot{
		AppData:   pb.AppData,
		EpochData: EpochDataFromPb(pb.EpochData),
	}
}

func (m *StateSnapshot) Pb() *commonpb.StateSnapshot {
	return &commonpb.StateSnapshot{
		AppData:   m.AppData,
		EpochData: (m.EpochData).Pb(),
	}
}

func (*StateSnapshot) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.StateSnapshot]()}
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
	EpochNr     types.EpochNr
	FirstSn     types.SeqNr
	Length      uint64
	Memberships []*Membership
}

func EpochConfigFromPb(pb *commonpb.EpochConfig) *EpochConfig {
	return &EpochConfig{
		EpochNr: (types.EpochNr)(pb.EpochNr),
		FirstSn: (types.SeqNr)(pb.FirstSn),
		Length:  pb.Length,
		Memberships: types1.ConvertSlice(pb.Memberships, func(t *commonpb.Membership) *Membership {
			return MembershipFromPb(t)
		}),
	}
}

func (m *EpochConfig) Pb() *commonpb.EpochConfig {
	return &commonpb.EpochConfig{
		EpochNr: (uint64)(m.EpochNr),
		FirstSn: (uint64)(m.FirstSn),
		Length:  m.Length,
		Memberships: types1.ConvertSlice(m.Memberships, func(t *Membership) *commonpb.Membership {
			return (t).Pb()
		}),
	}
}

func (*EpochConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.EpochConfig]()}
}

type Membership struct {
	Membership map[types2.NodeID]string
}

func MembershipFromPb(pb *commonpb.Membership) *Membership {
	return &Membership{
		Membership: types1.ConvertMap(pb.Membership, func(k string, v string) (types2.NodeID, string) {
			return (types2.NodeID)(k), v
		}),
	}
}

func (m *Membership) Pb() *commonpb.Membership {
	return &commonpb.Membership{
		Membership: types1.ConvertMap(m.Membership, func(k types2.NodeID, v string) (string, string) {
			return (string)(k), v
		}),
	}
}

func (*Membership) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.Membership]()}
}

type ClientProgress struct {
	Progress map[string]*DeliveredReqs
}

func ClientProgressFromPb(pb *commonpb.ClientProgress) *ClientProgress {
	return &ClientProgress{
		Progress: types1.ConvertMap(pb.Progress, func(k string, v *commonpb.DeliveredReqs) (string, *DeliveredReqs) {
			return k, DeliveredReqsFromPb(v)
		}),
	}
}

func (m *ClientProgress) Pb() *commonpb.ClientProgress {
	return &commonpb.ClientProgress{
		Progress: types1.ConvertMap(m.Progress, func(k string, v *DeliveredReqs) (string, *commonpb.DeliveredReqs) {
			return k, (v).Pb()
		}),
	}
}

func (*ClientProgress) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.ClientProgress]()}
}

type DeliveredReqs struct {
	LowWm     uint64
	Delivered []uint64
}

func DeliveredReqsFromPb(pb *commonpb.DeliveredReqs) *DeliveredReqs {
	return &DeliveredReqs{
		LowWm:     pb.LowWm,
		Delivered: pb.Delivered,
	}
}

func (m *DeliveredReqs) Pb() *commonpb.DeliveredReqs {
	return &commonpb.DeliveredReqs{
		LowWm:     m.LowWm,
		Delivered: m.Delivered,
	}
}

func (*DeliveredReqs) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.DeliveredReqs]()}
}
