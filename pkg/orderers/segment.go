package orderers

import (
	"github.com/filecoin-project/mir/pkg/orderers/types"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	ordererpbtypes "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// The Segment type represents an ISS Segment.
// It is used to parametrize an orderer (i.e. the SB instance).
type Segment ordererpbtypes.PBFTSegment

func NewSegment(leader t.NodeID, membership *commonpbtypes.Membership, proposals map[tt.SeqNr][]byte) *Segment {
	return (*Segment)(&ordererpbtypes.PBFTSegment{
		Leader:     leader,
		Membership: membership,
		Proposals:  proposals,
	})
}

func (seg *Segment) PbType() *ordererpbtypes.PBFTSegment {
	return (*ordererpbtypes.PBFTSegment)(seg)
}

func (seg *Segment) Len() int {
	return len(seg.Proposals)
}

func (seg *Segment) NodeIDs() []t.NodeID {
	return maputil.GetSortedKeys(seg.Membership.Nodes)
}

func (seg *Segment) PrimaryNode(view types.ViewNr) t.NodeID {
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

func (seg *Segment) SeqNrs() []tt.SeqNr {
	return maputil.GetSortedKeys(seg.Proposals)
}
