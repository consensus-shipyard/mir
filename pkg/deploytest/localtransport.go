package deploytest

import (
	"time"

	"github.com/filecoin-project/mir/pkg/trantor/types"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/testsim"
	t "github.com/filecoin-project/mir/pkg/types"
)

type LocalTransportLayer interface {
	Link(source t.NodeID) (net.Transport, error)
	Membership() *trantorpbtypes.Membership
	Close()
}

// NewLocalTransportLayer creates an instance of LocalTransportLayer suitable for tests.
// transportType is one of: "sim", "fake", "grpc", or "libp2p".
func NewLocalTransportLayer(sim *Simulation, transportType string, nodeIDsWeight map[t.NodeID]types.VoteWeight, logger logging.Logger) (LocalTransportLayer, error) {
	switch transportType {
	case "sim":
		messageDelayFn := func(from, to t.NodeID) time.Duration {
			// TODO: Make min and max message delay configurable
			return testsim.RandDuration(sim.Rand, 0, 10*time.Millisecond)
		}
		return NewSimTransport(sim, nodeIDsWeight, messageDelayFn), nil
	case "fake":
		return NewFakeTransport(nodeIDsWeight), nil
	case "grpc":
		return NewLocalGrpcTransport(nodeIDsWeight, logger)
	case "libp2p":
		return NewLocalLibp2pTransport(nodeIDsWeight, logger), nil
	default:
		return nil, es.Errorf("unexpected transport type: %v", transportType)
	}
}
