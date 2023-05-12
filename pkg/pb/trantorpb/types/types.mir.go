package trantorpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	trantorpb "github.com/filecoin-project/mir/pkg/pb/trantorpb"
	types "github.com/filecoin-project/mir/pkg/trantor/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Transaction struct {
	ClientId types.ClientID
	TxNo     types.TxNo
	Type     uint64
	Data     []uint8
}

func TransactionFromPb(pb *trantorpb.Transaction) *Transaction {
	if pb == nil {
		return nil
	}
	return &Transaction{
		ClientId: (types.ClientID)(pb.ClientId),
		TxNo:     (types.TxNo)(pb.TxNo),
		Type:     pb.Type,
		Data:     pb.Data,
	}
}

func (m *Transaction) Pb() *trantorpb.Transaction {
	if m == nil {
		return nil
	}
	return &trantorpb.Transaction{
		ClientId: (string)(m.ClientId),
		TxNo:     (uint64)(m.TxNo),
		Type:     m.Type,
		Data:     m.Data,
	}
}

func (*Transaction) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*trantorpb.Transaction]()}
}

type StateSnapshot struct {
	AppData   []uint8
	EpochData *EpochData
}

func StateSnapshotFromPb(pb *trantorpb.StateSnapshot) *StateSnapshot {
	if pb == nil {
		return nil
	}
	return &StateSnapshot{
		AppData:   pb.AppData,
		EpochData: EpochDataFromPb(pb.EpochData),
	}
}

func (m *StateSnapshot) Pb() *trantorpb.StateSnapshot {
	if m == nil {
		return nil
	}
	return &trantorpb.StateSnapshot{
		AppData:   m.AppData,
		EpochData: (m.EpochData).Pb(),
	}
}

func (*StateSnapshot) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*trantorpb.StateSnapshot]()}
}

type EpochData struct {
	EpochConfig        *EpochConfig
	ClientProgress     *ClientProgress
	LeaderPolicy       []uint8
	PreviousMembership *Membership
}

func EpochDataFromPb(pb *trantorpb.EpochData) *EpochData {
	if pb == nil {
		return nil
	}
	return &EpochData{
		EpochConfig:        EpochConfigFromPb(pb.EpochConfig),
		ClientProgress:     ClientProgressFromPb(pb.ClientProgress),
		LeaderPolicy:       pb.LeaderPolicy,
		PreviousMembership: MembershipFromPb(pb.PreviousMembership),
	}
}

func (m *EpochData) Pb() *trantorpb.EpochData {
	if m == nil {
		return nil
	}
	return &trantorpb.EpochData{
		EpochConfig:        (m.EpochConfig).Pb(),
		ClientProgress:     (m.ClientProgress).Pb(),
		LeaderPolicy:       m.LeaderPolicy,
		PreviousMembership: (m.PreviousMembership).Pb(),
	}
}

func (*EpochData) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*trantorpb.EpochData]()}
}

type EpochConfig struct {
	EpochNr     types.EpochNr
	FirstSn     types.SeqNr
	Length      uint64
	Memberships []*Membership
}

func EpochConfigFromPb(pb *trantorpb.EpochConfig) *EpochConfig {
	if pb == nil {
		return nil
	}
	return &EpochConfig{
		EpochNr: (types.EpochNr)(pb.EpochNr),
		FirstSn: (types.SeqNr)(pb.FirstSn),
		Length:  pb.Length,
		Memberships: types1.ConvertSlice(pb.Memberships, func(t *trantorpb.Membership) *Membership {
			return MembershipFromPb(t)
		}),
	}
}

func (m *EpochConfig) Pb() *trantorpb.EpochConfig {
	if m == nil {
		return nil
	}
	return &trantorpb.EpochConfig{
		EpochNr: (uint64)(m.EpochNr),
		FirstSn: (uint64)(m.FirstSn),
		Length:  m.Length,
		Memberships: types1.ConvertSlice(m.Memberships, func(t *Membership) *trantorpb.Membership {
			return (t).Pb()
		}),
	}
}

func (*EpochConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*trantorpb.EpochConfig]()}
}

type Membership struct {
	Nodes map[types2.NodeID]*NodeIdentity
}

func MembershipFromPb(pb *trantorpb.Membership) *Membership {
	if pb == nil {
		return nil
	}
	return &Membership{
		Nodes: types1.ConvertMap(pb.Nodes, func(k string, v *trantorpb.NodeIdentity) (types2.NodeID, *NodeIdentity) {
			return (types2.NodeID)(k), NodeIdentityFromPb(v)
		}),
	}
}

func (m *Membership) Pb() *trantorpb.Membership {
	if m == nil {
		return nil
	}
	return &trantorpb.Membership{
		Nodes: types1.ConvertMap(m.Nodes, func(k types2.NodeID, v *NodeIdentity) (string, *trantorpb.NodeIdentity) {
			return (string)(k), (v).Pb()
		}),
	}
}

func (*Membership) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*trantorpb.Membership]()}
}

type NodeIdentity struct {
	Id     types2.NodeID
	Addr   string
	Key    []uint8
	Weight uint64
}

func NodeIdentityFromPb(pb *trantorpb.NodeIdentity) *NodeIdentity {
	if pb == nil {
		return nil
	}
	return &NodeIdentity{
		Id:     (types2.NodeID)(pb.Id),
		Addr:   pb.Addr,
		Key:    pb.Key,
		Weight: pb.Weight,
	}
}

func (m *NodeIdentity) Pb() *trantorpb.NodeIdentity {
	if m == nil {
		return nil
	}
	return &trantorpb.NodeIdentity{
		Id:     (string)(m.Id),
		Addr:   m.Addr,
		Key:    m.Key,
		Weight: m.Weight,
	}
}

func (*NodeIdentity) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*trantorpb.NodeIdentity]()}
}

type ClientProgress struct {
	Progress map[types.ClientID]*DeliveredTXs
}

func ClientProgressFromPb(pb *trantorpb.ClientProgress) *ClientProgress {
	if pb == nil {
		return nil
	}
	return &ClientProgress{
		Progress: types1.ConvertMap(pb.Progress, func(k string, v *trantorpb.DeliveredTXs) (types.ClientID, *DeliveredTXs) {
			return (types.ClientID)(k), DeliveredTXsFromPb(v)
		}),
	}
}

func (m *ClientProgress) Pb() *trantorpb.ClientProgress {
	if m == nil {
		return nil
	}
	return &trantorpb.ClientProgress{
		Progress: types1.ConvertMap(m.Progress, func(k types.ClientID, v *DeliveredTXs) (string, *trantorpb.DeliveredTXs) {
			return (string)(k), (v).Pb()
		}),
	}
}

func (*ClientProgress) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*trantorpb.ClientProgress]()}
}

type DeliveredTXs struct {
	LowWm     uint64
	Delivered []uint64
}

func DeliveredTXsFromPb(pb *trantorpb.DeliveredTXs) *DeliveredTXs {
	if pb == nil {
		return nil
	}
	return &DeliveredTXs{
		LowWm:     pb.LowWm,
		Delivered: pb.Delivered,
	}
}

func (m *DeliveredTXs) Pb() *trantorpb.DeliveredTXs {
	if m == nil {
		return nil
	}
	return &trantorpb.DeliveredTXs{
		LowWm:     m.LowWm,
		Delivered: m.Delivered,
	}
}

func (*DeliveredTXs) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*trantorpb.DeliveredTXs]()}
}
