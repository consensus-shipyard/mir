package deploytest

import (
	"fmt"

	"github.com/multiformats/go-multiaddr"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

var _ LocalTransportLayer = &LocalGrpcTransport{}

type LocalGrpcTransport struct {
	// Complete static membership of the system.
	// Maps the node ID of each node in the system to a string representation of its network address.
	// The address format "IPAddress:port"
	membership *commonpbtypes.Membership

	// Logger is used for all logging events of this LocalGrpcTransport
	logger logging.Logger
}

func NewLocalGrpcTransport(nodeIDs []t.NodeID, logger logging.Logger) *LocalGrpcTransport {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	// Compute network addresses and ports for all test replicas.
	// Each test replica is on the local machine - 127.0.0.1
	membership := &commonpbtypes.Membership{make(map[t.NodeID]*commonpbtypes.NodeIdentity)} // nolint:govet
	for i, id := range nodeIDs {
		maddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", BaseListenPort+i))
		if err != nil {
			panic(err)
		}
		membership.Nodes[id] = &commonpbtypes.NodeIdentity{id, maddr.String(), nil, 0} // nolint:govet
	}

	return &LocalGrpcTransport{membership, logger}
}

func (t *LocalGrpcTransport) Link(sourceID t.NodeID) (net.Transport, error) {
	return grpc.NewTransport(
		sourceID,
		t.membership.Nodes[sourceID].Addr,
		logging.Decorate(t.logger, fmt.Sprintf("gRPC: Node %v: ", sourceID)),
	)
}

func (t *LocalGrpcTransport) Membership() *commonpbtypes.Membership {
	return t.membership
}

func (t *LocalGrpcTransport) Close() {}
