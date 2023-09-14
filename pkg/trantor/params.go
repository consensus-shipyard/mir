package trantor

import (
	"time"

	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
)

type Params struct {
	Mempool      *simplemempool.ModuleParams
	Iss          *issconfig.ModuleParams
	Net          libp2p.Params
	Availability multisigcollector.ModuleParams
}

func DefaultParams(initialMembership *trantorpbtypes.Membership) Params {
	return Params{
		Mempool:      simplemempool.DefaultModuleParams(),
		Iss:          issconfig.DefaultParams(initialMembership),
		Net:          libp2p.DefaultParams(),
		Availability: multisigcollector.DefaultParamsTemplate(),
	}
}

func (p *Params) AdjustSpeed(maxProposeDelay time.Duration) *Params {
	p.Iss.AdjustSpeed(maxProposeDelay)
	p.Mempool.BatchTimeout = maxProposeDelay // TODO: account for processing time
	return p
}
