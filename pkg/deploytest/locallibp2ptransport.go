package deploytest

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	t "github.com/filecoin-project/mir/pkg/types"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
	"strconv"
)

type LocalLibp2pTransport struct {
	// Complete static membership of the system.
	// Maps the node ID of each node in the system to a string representation of its libp2p address.
	membership map[t.NodeID]string

	// Logger is used for all logging events of this LocalGrpcTransport
	logger logging.Logger
}

func NewLocalLibp2pTransport(nodeIDs []t.NodeID, logger logging.Logger) *LocalLibp2pTransport {
	membership := make(map[t.NodeID]string, len(nodeIDs))
	for i, id := range nodeIDs {
		membership[id] = libp2ptools.NewDummyHostID(i, BaseListenPort).String()
	}

	return &LocalLibp2pTransport{membership, logger}
}

func (t *LocalLibp2pTransport) Link(sourceID t.NodeID) net.Transport {
	// TODO: it should be possible for sourceID to be an arbitrary string
	id, err := strconv.Atoi(string(sourceID))
	if err != nil {
		panic(err)
	}

	h := libp2ptools.NewDummyHost(id, BaseListenPort)

	return libp2p.NewTransport(
		h,
		t.membership,
		sourceID,
		t.logger,
	)
}
