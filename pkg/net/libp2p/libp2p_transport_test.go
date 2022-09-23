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

type mockLibp2pTransport struct {
	t          *testing.T
	membership map[types.NodeID]types.NodeAddress
	hosts      map[types.NodeID]host.Host
	logger     logging.Logger
}

func newMockLibp2pTransport(t *testing.T, nodeIDs []types.NodeID, logger logging.Logger, port int) *mockLibp2pTransport {
	lt := &mockLibp2pTransport{
		t:          t,
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

func (m *mockLibp2pTransport) NodeTransport(sourceID types.NodeID) (*Transport, error) {
	if _, ok := m.hosts[sourceID]; !ok {
		panic(fmt.Errorf("unexpected node id: %v", sourceID))
	}

	return NewTransport(
		m.hosts[sourceID],
		sourceID,
		m.logger,
	)
}

func (m *mockLibp2pTransport) CloseConnection(srcNode, dstNode types.NodeID) {
	src, err := m.NodeTransport(srcNode)
	require.NoError(m.t, err)
	dst, err := m.NodeTransport(dstNode)
	require.NoError(m.t, err)

	err = src.host.Network().ClosePeer(dst.host.ID())
	require.NoError(m.t, err)

	err = dst.host.Network().ClosePeer(src.host.ID())
	require.NoError(m.t, err)
}

func (m *mockLibp2pTransport) Nodes() map[types.NodeID]types.NodeAddress {
	return m.membership
}

func TestLibp2pReconnect(t *testing.T) {
	ctx := context.Background()
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")

	m := newMockLibp2pTransport(t, []types.NodeID{nodeA, nodeB, nodeC}, logger, 10000)

	a, err := m.NodeTransport(nodeA)
	require.NoError(t, err)
	err = a.Start()
	require.NoError(t, err)
	defer a.Stop()

	b, err := m.NodeTransport(nodeB)
	require.NoError(t, err)
	err = b.Start()
	require.NoError(t, err)
	defer b.Stop()

	c, err := m.NodeTransport(nodeC)
	require.NoError(t, err)
	err = c.Start()
	require.NoError(t, err)
	defer c.Stop()

	a.syncConnect(ctx, m.Nodes())
	b.syncConnect(ctx, m.Nodes())
	c.syncConnect(ctx, m.Nodes())

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

	m.CloseConnection(nodeA, nodeB)

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
