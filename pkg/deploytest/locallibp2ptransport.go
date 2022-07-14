package deploytest

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	t "github.com/filecoin-project/mir/pkg/types"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/multiformats/go-multiaddr"
)

type LocalLibp2pTransport struct {
	// Complete static membership of the system.
	// Maps the node ID of each node in the system to its libp2p address.
	membership map[t.NodeID]multiaddr.Multiaddr

	// Maps node ids to their libp2p host.
	hosts map[t.NodeID]host.Host

	// Logger is used for all logging events of this LocalGrpcTransport
	logger logging.Logger
}

func NewLocalLibp2pTransport(nodeIDs []t.NodeID, logger logging.Logger) *LocalLibp2pTransport {
	lt := &LocalLibp2pTransport{
		membership: make(map[t.NodeID]multiaddr.Multiaddr, len(nodeIDs)),
		hosts:      make(map[t.NodeID]host.Host),
		logger:     logger,
	}

	for i, id := range nodeIDs {
		lt.hosts[id] = libp2ptools.NewDummyHost(i, BaseListenPort)
		lt.membership[id] = libp2ptools.NewDummyPeerID(i, BaseListenPort)
	}

	return lt
}

func (lt *LocalLibp2pTransport) Link(sourceID t.NodeID) net.Transport {
	if _, ok := lt.hosts[sourceID]; !ok {
		panic(fmt.Errorf("unexpected node id: %v", sourceID))
	}

	return libp2p.NewTransport(
		lt.hosts[sourceID],
		lt.membership,
		sourceID,
		lt.logger,
	)
}
