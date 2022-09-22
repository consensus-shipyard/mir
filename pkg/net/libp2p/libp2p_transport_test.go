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
	nodeC := types.NodeID("c")

	h := newLibp2pTransportHarness([]types.NodeID{nodeA, nodeB, nodeC}, logger, 10000)

	a, err := h.Link(nodeA)
	require.NoError(t, err)
	err = a.Start()
	defer a.Stop()
	require.NoError(t, err)

	b, err := h.Link(nodeB)
	require.NoError(t, err)
	err = b.Start()
	defer b.Stop()
	require.NoError(t, err)

	c, err := h.Link(nodeC)
	require.NoError(t, err)
	err = c.Start()
	defer c.Stop()
	require.NoError(t, err)

	a.syncConnect(ctx, h.Nodes())
	b.syncConnect(ctx, h.Nodes())
	c.syncConnect(ctx, h.Nodes())

	nodeBEventsChan := b.EventsOut()
	nodeCEventsChan := c.EventsOut()

	err = a.Send(nodeB, &messagepb.Message{})
	require.NoError(t, err)
	nodeBEvents := <-nodeBEventsChan
	msg, valid := nodeBEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())

	err = a.Send(nodeC, &messagepb.Message{})
	require.NoError(t, err)
	nodeCEvents := <-nodeCEventsChan
	msg, valid = nodeCEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())

	err = b.host.Network().ClosePeer(a.host.ID())
	require.NoError(t, err)

	n := len(b.host.Network().Peers())
	require.Equal(t, 1, n)

	time.Sleep(3 * time.Second)

	err = a.Send(nodeB, &messagepb.Message{})
	require.Equal(t, mirnet.ErrWritingFailed, err)

	err = a.Send(nodeC, &messagepb.Message{})
	require.NoError(t, err)
	nodeCEvents = <-nodeCEventsChan
	msg, valid = nodeCEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())

	err = a.Send(nodeB, &messagepb.Message{})
	require.NoError(t, err)
	nodeBEvents = <-nodeBEventsChan
	msg, valid = nodeBEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())
}
