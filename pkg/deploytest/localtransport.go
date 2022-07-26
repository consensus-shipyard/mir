package deploytest

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	t "github.com/filecoin-project/mir/pkg/types"
)

type LocalTransportLayer interface {
	Link(source t.NodeID) (net.Transport, error)
	Nodes() map[t.NodeID]t.NodeAddress
}

// NewLocalTransportLayer creates an instance of LocalTransportLayer suitable for tests.
// transportType is one of: "fake", "grpc", or "libp2p".
func NewLocalTransportLayer(transportType string, nodeIDs []t.NodeID, logger logging.Logger) LocalTransportLayer {
	switch transportType {
	case "fake":
		return NewFakeTransport(nodeIDs)
	case "grpc":
		return NewLocalGrpcTransport(nodeIDs, logger)
	case "libp2p":
		return NewLocalLibp2pTransport(nodeIDs, logger)
	default:
		panic(fmt.Sprintf("unexpected transport type: %v", transportType))
	}
}
