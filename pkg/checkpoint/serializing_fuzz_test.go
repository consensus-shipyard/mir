// This is just a "Hello World" example of fuzz testing. It is not meant to be exhaustive.

package checkpoint

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
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
		membership := make(map[string]string)

		for i := 0; i < n; i++ {
			membership[fmt.Sprintf("%s/%s", k, strconv.Itoa(i))] = fmt.Sprintf("%s%s", v, strconv.Itoa(i))
		}

		mb := commonpb.Membership{Membership: membership}
		cfg := commonpb.EpochConfig{EpochNr: e, Memberships: []*commonpb.Membership{&mb}}
		clProgress := commonpb.ClientProgress{Progress: map[string]*commonpb.DeliveredReqs{}} // TODO: add actual values
		state := commonpb.StateSnapshot{AppData: data, EpochData: &commonpb.EpochData{
			EpochConfig:        &cfg,
			ClientProgress:     &clProgress,
			PreviousMembership: types.MembershipPb(make(map[types.NodeID]types.NodeAddress)),
		}}
		serializeSnapshotForHash(&state)
	})
}
