package serializing

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	"github.com/filecoin-project/mir/pkg/types"
)

func FuzzRequestForHash(f *testing.F) {
	f.Add("/ip4/7.7.7.7/tcp/1234/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N", 0, 1, []byte("13242342342342"))

	f.Fuzz(func(t *testing.T, id string, tp, n int, data []byte) {
		req := requestpb.Request{ClientId: id, ReqNo: uint64(n), Type: uint64(tp), Data: data}
		RequestForHash(&req)
	})
}

func FuzzCheckpointForSig(f *testing.F) {
	f.Add(uint64(0), uint64(0), []byte("13242342342342"))

	f.Fuzz(func(t *testing.T, s, n uint64, data []byte) {
		CheckpointForSig(types.EpochNr(s), types.SeqNr(n), data)
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
			EpochConfig:    &cfg,
			ClientProgress: &clProgress,
		}}
		SnapshotForHash(&state)
	})
}
