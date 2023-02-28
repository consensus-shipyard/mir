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

	// Sequence numbers for which the orderer is responsible, along with corresponding (optional) pre-defined proposals.
	// The keys of this map are the actual "segment" of the commit log.
	// A nil value means that no proposal is specified (and the protocol implementation will decide what to propose).
	// A non-nil value will be proposed (by this node) for that sequence number whenever possible.
	// Currently, such a "free" proposal is a new availability certificate in view 0,
	// and a special empty one in other views.
	Proposals map[t.SeqNr][]byte
}

func NewSegment(leader t.NodeID, membership map[t.NodeID]t.NodeAddress, proposals map[t.SeqNr][]byte) *Segment {
	return &Segment{
		Leader:     leader,
		Membership: membership,
		Proposals:  proposals,
	}
}

func SegmentFromPb(seg *ordererspb.PBFTSegment) *Segment {
	return &Segment{
		Leader:     t.NodeID(seg.Leader),
		Membership: t.Membership(seg.Membership),
		Proposals: maputil.Transform(
			seg.Proposals,
			func(key uint64, val []byte) (t.SeqNr, []byte) {
				return t.SeqNr(key), val
			},
		),
	}
}

func (seg *Segment) Pb() *ordererspb.PBFTSegment {
	return &ordererspb.PBFTSegment{
		Leader:     seg.Leader.Pb(),
		Membership: t.MembershipPb(seg.Membership),
		Proposals: maputil.Transform(
			seg.Proposals,
			func(key t.SeqNr, val []byte) (uint64, []byte) {
				return key.Pb(), val
			},
		),
	}
}

func (seg *Segment) Len() int {
	return len(seg.Proposals)
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

func (seg *Segment) SeqNrs() []t.SeqNr {
	return maputil.GetSortedKeys(seg.Proposals)
}
