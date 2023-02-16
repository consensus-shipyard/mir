package orderers

import (
	"github.com/filecoin-project/mir/pkg/pb/ordererspb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// The Segment type represents an ISS Segment.
// It is used to parametrize an orderer (i.e. the SB instance).
type Segment struct {

	// The leader node of the orderer.
	Leader t.NodeID

	// All nodes executing the orderer implementation.
	Membership map[t.NodeID]t.NodeAddress

	// List of sequence numbers for which the orderer is responsible.
	// This is the actual "segment" of the commit log.
	SeqNrs []t.SeqNr

	// List of proposals, one per sequence number.
	// A non-nil proposal will always be proposed by a node for the corresponding sequence number, regardless of view
	// (unless, of course, a previous proposal is enforced by the protocol).
	// If a proposal is nil, the implementation is left to choose the proposal.
	// Currently, such a "free" proposal is a new availability certificate in view 0,
	// and a special empty one in other views.
	Proposals [][]byte
}

func NewSegment(leader t.NodeID, membership map[t.NodeID]t.NodeAddress, seqNrs []t.SeqNr, proposals [][]byte) *Segment {
	return &Segment{
		Leader:     leader,
		Membership: membership,
		SeqNrs:     seqNrs,
		Proposals:  proposals,
	}
}

func SegmentFromPb(seg *ordererspb.PBFTSegment) *Segment {
	return &Segment{
		Leader:     t.NodeID(seg.Leader),
		Membership: t.Membership(seg.Membership),
		SeqNrs:     t.SeqNrSlice(seg.SeqNrs),
		Proposals:  seg.Proposals,
	}
}

func (seg *Segment) NodeIDs() []t.NodeID {
	return maputil.GetSortedKeys(seg.Membership)
}

func (seg *Segment) PrimaryNode(view t.PBFTViewNr) t.NodeID {
	return seg.NodeIDs()[(seg.LeaderIndex()+int(view))%len(seg.NodeIDs())]
}

func (seg *Segment) LeaderIndex() int {
	for i, nodeID := range seg.NodeIDs() {
		if nodeID == seg.Leader {
			return i
		}
	}
	panic("invalid segment: leader not in membership")
}

func (seg *Segment) Pb() *ordererspb.PBFTSegment {
	return &ordererspb.PBFTSegment{
		Leader:     seg.Leader.Pb(),
		Membership: t.MembershipPb(seg.Membership),
		SeqNrs:     t.SeqNrSlicePb(seg.SeqNrs),
		Proposals:  seg.Proposals,
	}
}
