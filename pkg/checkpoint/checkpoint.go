package checkpoint

import (
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type StableCheckpoint checkpointpb.StableCheckpoint

func (sc *StableCheckpoint) SeqNr() t.SeqNr {
	return t.SeqNr(sc.Sn)
}

func (sc *StableCheckpoint) Memberships() []map[t.NodeID]t.NodeAddress {
	return t.MembershipSlice(sc.Snapshot.Configuration.Memberships)
}

func (sc *StableCheckpoint) Epoch() t.EpochNr {
	return t.EpochNr(sc.Snapshot.Configuration.EpochNr)
}

func (sc *StableCheckpoint) StateSnapshot() *commonpb.StateSnapshot {
	return sc.Snapshot
}

func (sc *StableCheckpoint) Pb() *checkpointpb.StableCheckpoint {
	return (*checkpointpb.StableCheckpoint)(sc)
}

func Genesis(initialStateSnapshot *commonpb.StateSnapshot) *StableCheckpoint {
	return &StableCheckpoint{
		Sn:       0,
		Snapshot: initialStateSnapshot,
		Cert:     map[string][]byte{},
	}
}
