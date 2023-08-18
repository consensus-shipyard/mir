package deploytest

import (
	es "github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/filecoin-project/mir/pkg/trantor/types"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
)

type LocalLibp2pTransport struct {
	// Complete static membership of the system.
	// Maps the node ID of each node in the system to its libp2p address.
	membership *trantorpbtypes.Membership

	// Maps node ids to their libp2p host.
	hosts map[t.NodeID]host.Host

	// Logger is used for all logging events of this LocalGrpcTransport
	logger logging.Logger
}

func NewLocalLibp2pTransport(nodeIDsWeight map[t.NodeID]types.VoteWeight, logger logging.Logger) *LocalLibp2pTransport {
	lt := &LocalLibp2pTransport{
		membership: &trantorpbtypes.Membership{ // nolint:govet
			make(map[t.NodeID]*trantorpbtypes.NodeIdentity, len(nodeIDsWeight)),
		},
		hosts:  make(map[t.NodeID]host.Host),
		logger: logger,
	}

	i := 0
	for id, weight := range nodeIDsWeight {
		hostAddr := libp2ptools.NewDummyHostAddr(i, BaseListenPort)
		lt.hosts[id] = libp2ptools.NewDummyHost(i, hostAddr)
		lt.membership.Nodes[id] = &trantorpbtypes.NodeIdentity{ // nolint:govet
			id,
			libp2ptools.NewDummyMultiaddr(i, hostAddr).String(),
			nil,
			weight,
		}
		i++
	}

	return lt
}

func (t *LocalLibp2pTransport) Link(sourceID t.NodeID) (net.Transport, error) {
	if _, ok := t.hosts[sourceID]; !ok {
		return nil, es.Errorf("unexpected node id: %v", sourceID)
	}

	return libp2p.NewTransport(
		libp2p.DefaultParams(),
		sourceID,
		t.hosts[sourceID],
		t.logger,
		nil,
	), nil
}

func (t *LocalLibp2pTransport) Membership() *trantorpbtypes.Membership {
	return t.membership
}

func (t *LocalLibp2pTransport) Close() {
	for _, h := range t.hosts {
		h.Close() // nolint
	}
}
