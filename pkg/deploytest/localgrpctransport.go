package deploytest

import (
	"fmt"

	"github.com/multiformats/go-multiaddr"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	t "github.com/filecoin-project/mir/pkg/types"
)

type LocalGrpcTransport struct {
	// Complete static membership of the system.
	// Maps the node ID of each node in the system to a string representation of its network address.
	// The address format "IPAddress:port"
	membership map[t.NodeID]t.NodeAddress // nodeId -> "IPAddress:port"

	// Logger is used for all logging events of this LocalGrpcTransport
	logger logging.Logger
}

func NewLocalGrpcTransport(nodeIDs []t.NodeID, logger logging.Logger) *LocalGrpcTransport {
	// Compute network addresses and ports for all test replicas.
	// Each test replica is on the local machine - 127.0.0.1
	membership := make(map[t.NodeID]t.NodeAddress)
	for i, id := range nodeIDs {
		maddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", BaseListenPort+i))
		if err != nil {
			panic(err)
		}
		membership[id] = t.NodeAddress(maddr)
	}

	return &LocalGrpcTransport{membership, logger}
}

func (t *LocalGrpcTransport) Link(sourceID t.NodeID) (net.Transport, error) {
	return grpc.NewTransport(
		sourceID,
		t.membership[sourceID],
		logging.Decorate(t.logger, fmt.Sprintf("gRPC: Node %v: ", sourceID)),
	)
}

func (t *LocalGrpcTransport) Nodes() map[t.NodeID]t.NodeAddress {
	return t.membership
}

func (t *LocalGrpcTransport) Close() {}
