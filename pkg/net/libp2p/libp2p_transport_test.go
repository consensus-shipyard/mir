package libp2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/mir/pkg/logging"
	mirnet "github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
	tools "github.com/filecoin-project/mir/pkg/util/libp2p"
)

type libp2pTransportHarness struct {
	// Complete static membership of the system.
	// Maps the node ID of each node in the system to its libp2p address.
	membership map[types.NodeID]types.NodeAddress

	// Maps node ids to their libp2p host.
	hosts map[types.NodeID]host.Host

	// Logger is used for all logging events of this LocalGrpcTransport
	logger logging.Logger
}

func newLibp2pTransportHarness(nodeIDs []types.NodeID, logger logging.Logger, port int) *libp2pTransportHarness {
	lt := &libp2pTransportHarness{
		membership: make(map[types.NodeID]types.NodeAddress, len(nodeIDs)),
		hosts:      make(map[types.NodeID]host.Host),
		logger:     logger,
	}

	for i, id := range nodeIDs {
		hostAddr := tools.NewDummyHostAddr(i, port)
		lt.hosts[id] = tools.NewDummyHost(i, hostAddr)
		lt.membership[id] = types.NodeAddress(tools.NewDummyMultiaddr(i, hostAddr))
	}

	return lt
}

func (t *libp2pTransportHarness) Link(sourceID types.NodeID) (*Transport, error) {
	if _, ok := t.hosts[sourceID]; !ok {
		panic(fmt.Errorf("unexpected node id: %v", sourceID))
	}

	return NewTransport(
		t.hosts[sourceID],
		sourceID,
		t.logger,
	)
}

func (t *libp2pTransportHarness) Nodes() map[types.NodeID]types.NodeAddress {
	return t.membership
}

func TestLibp2pReconnect(t *testing.T) {
	ctx := context.Background()
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeIDs := []types.NodeID{nodeA, nodeB}

	h := newLibp2pTransportHarness(nodeIDs, logger, 10000)

	a, err := h.Link(nodeIDs[0])
	require.NoError(t, err)
	err = a.Start()
	require.NoError(t, err)

	b, err := h.Link(nodeIDs[1])
	require.NoError(t, err)
	err = b.Start()
	require.NoError(t, err)

	a.syncConnect(ctx, h.Nodes())

	nodeBEventsChan := b.EventsOut()

	err = a.Send(nodeIDs[1], &messagepb.Message{})
	require.NoError(t, err)

	events := <-nodeBEventsChan
	e := events.Iterator().Next()
	msg, valid := e.Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())

	err = b.host.Network().ClosePeer(a.host.ID())
	require.NoError(t, err)

	n := len(b.host.Network().Peers())
	require.Equal(t, 0, n)

	time.Sleep(3 * time.Second)

	err = a.Send(nodeIDs[1], &messagepb.Message{})
	require.Equal(t, mirnet.ErrWritingFailed, err)

	err = a.Send(nodeIDs[1], &messagepb.Message{})
	require.NoError(t, err)

	events = <-nodeBEventsChan
	e = events.Iterator().Next()
	msg, valid = e.Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())

	a.Stop()
	b.Stop()
}
