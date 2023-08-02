// This is just a "Hello World" example of fuzz testing. It is not meant to be exhaustive.

package checkpoint

import (
	"fmt"
	"strconv"
	"testing"

	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/types"
)

func FuzzCheckpointForSig(f *testing.F) {
	f.Add(uint64(0), uint64(0), []byte("13242342342342"))

	f.Fuzz(func(t *testing.T, s, n uint64, data []byte) {
		serializeCheckpointForSig(tt.EpochNr(s), tt.SeqNr(n), data)
	})
}

func FuzzSnapshotForHash(f *testing.F) {
	f.Add(100, uint64(0), "/ip4/7.7.7.7/tcp/1234/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N", "127.0.0.1:3333", []byte("13242342342342"))

	f.Fuzz(func(t *testing.T, n int, e uint64, k, v string, data []byte) {
		n = n % 5000
		membership := trantorpbtypes.Membership{make(map[types.NodeID]*trantorpbtypes.NodeIdentity)} // nolint:govet

		for i := 0; i < n; i++ {
			id := types.NodeID(fmt.Sprintf("%s/%s", k, strconv.Itoa(i)))
			addr := fmt.Sprintf("%s%s", v, strconv.Itoa(i))
			membership.Nodes[id] = &trantorpbtypes.NodeIdentity{
				Id:     id,
				Addr:   addr,
				Key:    nil,
				Weight: "1",
			}
		}

		cfg := trantorpbtypes.EpochConfig{EpochNr: tt.EpochNr(e), Memberships: []*trantorpbtypes.Membership{&membership}}
		clProgress := trantorpbtypes.ClientProgress{Progress: map[tt.ClientID]*trantorpbtypes.DeliveredTXs{}} // TODO: add actual values
		state := trantorpbtypes.StateSnapshot{AppData: data, EpochData: &trantorpbtypes.EpochData{
			EpochConfig:    &cfg,
			ClientProgress: &clProgress,
			PreviousMembership: &trantorpbtypes.Membership{ // nolint:govet
				make(map[types.NodeID]*trantorpbtypes.NodeIdentity),
			},
		}}
		_, err := serializeSnapshotForHash(&state)
		if err != nil {
			panic(err)
		}
	})
}
