package smr

import (
	"time"

	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	t "github.com/filecoin-project/mir/pkg/types"
)

type Params struct {
	Mempool *simplemempool.ModuleParams
	Iss     *iss.ModuleParams
	Net     libp2p.Params
}

func DefaultParams(initialMembership map[t.NodeID]t.NodeAddress) Params {
	return Params{
		Mempool: simplemempool.DefaultModuleParams(),
		Iss:     iss.DefaultParams(initialMembership),
		Net:     libp2p.DefaultParams(),
	}
}

func (p *Params) AdjustSpeed(maxProposeDelay time.Duration) *Params {
	p.Iss.AdjustSpeed(maxProposeDelay)
	return p
}
