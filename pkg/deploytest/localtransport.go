package deploytest

import (
	"time"

	"github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/testsim"
)

type LocalTransportLayer interface {
	Link(source t.NodeID) (net.Transport, error)
	Membership() *trantorpbtypes.Membership
	Close()
}

// NewLocalTransportLayer creates an instance of LocalTransportLayer suitable for tests.
// transportType is one of: "sim", "fake", or "grpc".
func NewLocalTransportLayer(sim *Simulation, transportType string, nodeIDsWeight map[t.NodeID]types.VoteWeight, logger logging.Logger) (LocalTransportLayer, error) {
	switch transportType {
	case "sim":
		messageDelayFn := func(_, _ t.NodeID) time.Duration {
			// TODO: Make min and max message delay configurable
			return testsim.RandDuration(sim.Rand, 0, 10*time.Millisecond)
		}
		return NewSimTransport(sim, nodeIDsWeight, messageDelayFn), nil
	case "fake":
		return NewFakeTransport(nodeIDsWeight), nil
	case "grpc":
		return NewLocalGrpcTransport(nodeIDsWeight, logger)
	default:
		return nil, es.Errorf("unexpected transport type: %v", transportType)
	}
}
