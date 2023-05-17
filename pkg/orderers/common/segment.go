package common

import (
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/orderers/types"
	ordererpbtypes "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// The Segment type represents an ISS Segment.
// It is used to parametrize an orderer (i.e. the SB instance).
type Segment ordererpbtypes.PBFTSegment

func NewSegment(
	leader t.NodeID,
	membership *trantorpbtypes.Membership,
	proposals map[tt.SeqNr][]byte,
) (*Segment, error) {
	if _, ok := membership.Nodes[leader]; !ok {
		return nil, es.Errorf("leader (%v) not in Membership (%v)", leader, maputil.GetKeys(membership.Nodes))
	}

	return (*Segment)(&ordererpbtypes.PBFTSegment{
		Leader:     leader,
		Membership: membership,
		Proposals:  proposals,
	}), nil
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
	// Not returning an error here, since if we reach this line, there is an error in this very file.
	panic("invalid segment: leader not in Membership")
}

func (seg *Segment) SeqNrs() []tt.SeqNr {
	return maputil.GetSortedKeys(seg.Proposals)
}
