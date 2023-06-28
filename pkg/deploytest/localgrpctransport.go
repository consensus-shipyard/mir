package deploytest

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/trantor/types"

	es "github.com/go-errors/errors"
	"github.com/multiformats/go-multiaddr"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

var _ LocalTransportLayer = &LocalGrpcTransport{}

type LocalGrpcTransport struct {
	// Complete static membership of the system.
	// Maps the node ID of each node in the system to a string representation of its network address.
	// The address format "IPAddress:port"
	membership *trantorpbtypes.Membership

	// Logger is used for all logging events of this LocalGrpcTransport
	logger logging.Logger
}

func NewLocalGrpcTransport(nodeIDsWeight map[t.NodeID]types.VoteWeight, logger logging.Logger) (*LocalGrpcTransport, error) {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	// Compute network addresses and ports for all test replicas.
	// Each test replica is on the local machine - 127.0.0.1
	membership := &trantorpbtypes.Membership{make(map[t.NodeID]*trantorpbtypes.NodeIdentity)} // nolint:govet
	i := 0
	for id, weight := range nodeIDsWeight {
		maddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", BaseListenPort+i))
		if err != nil {
			return nil, es.Errorf("error creating local multiaddress: %w", err)
		}
		membership.Nodes[id] = &trantorpbtypes.NodeIdentity{id, maddr.String(), nil, weight} // nolint:govet
		i++
	}

	return &LocalGrpcTransport{membership, logger}, nil
}

func (t *LocalGrpcTransport) Link(sourceID t.NodeID) (net.Transport, error) {
	return grpc.NewTransport(
		sourceID,
		t.membership.Nodes[sourceID].Addr,
		logging.Decorate(t.logger, fmt.Sprintf("gRPC: Node %v: ", sourceID)),
	)
}

func (t *LocalGrpcTransport) Membership() *trantorpbtypes.Membership {
	return t.membership
}

func (t *LocalGrpcTransport) Close() {}
