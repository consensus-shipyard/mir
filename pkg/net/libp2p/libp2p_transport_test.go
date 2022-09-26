package libp2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
	tools "github.com/filecoin-project/mir/pkg/util/libp2p"
)

type mockLibp2pCommunication struct {
	t          *testing.T
	membership map[types.NodeID]types.NodeAddress
	hosts      map[types.NodeID]host.Host
	logger     logging.Logger
	transports map[types.NodeID]*Transport
}

func newMockLibp2pCommunication(t *testing.T, nodeIDs []types.NodeID, logger logging.Logger, port int) *mockLibp2pCommunication {
	lt := &mockLibp2pCommunication{
		t:          t,
		membership: make(map[types.NodeID]types.NodeAddress, len(nodeIDs)),
		hosts:      make(map[types.NodeID]host.Host),
		transports: make(map[types.NodeID]*Transport),
		logger:     logger,
	}

	var err error
	for i, id := range nodeIDs {
		hostAddr := tools.NewDummyHostAddr(i, port)
		lt.hosts[id] = tools.NewDummyHost(i, hostAddr)
		lt.membership[id] = types.NodeAddress(tools.NewDummyMultiaddr(i, hostAddr))
		lt.transports[id], err = NewTransport(lt.hosts[id], id, logger)
		if err != nil {
			t.Fatalf("failed to create transport for node %s: %v", id, err)
		}
	}

	return lt
}

func (m *mockLibp2pCommunication) getTransport(sourceID types.NodeID) *Transport {
	tr, ok := m.transports[sourceID]
	if !ok {
		m.t.Fatalf("failed to get transport for node id: %v", sourceID)
	}
	return tr
}

func (m *mockLibp2pCommunication) disconnect(srcNode, dstNode types.NodeID) {
	src := m.getTransport(srcNode)
	dst := m.getTransport(dstNode)

	err := src.host.Network().ClosePeer(dst.host.ID())
	require.NoError(m.t, err)

	err = dst.host.Network().ClosePeer(src.host.ID())
	require.NoError(m.t, err)
}

func (m *mockLibp2pCommunication) Nodes() map[types.NodeID]types.NodeAddress {
	return m.membership
}

func (m *mockLibp2pCommunication) sendEventually(ctx context.Context, srcNode, dstNode types.NodeID, msg *messagepb.Message) error {
	src := m.getTransport(srcNode)

	timeout := time.After(5 * time.Second)
	sendTicker := time.NewTicker(1400 * time.Millisecond)
	defer sendTicker.Stop()

	var err error
	for {
		select {
		case <-timeout:
			return err
		case <-sendTicker.C:
			err = src.Send(ctx, dstNode, msg)
			if err == nil {
				return nil
			}
			m.t.Logf("%s is trying to send message to %s: %v", srcNode, dstNode, err)
		}
	}
}

func TestLibp2pReconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")

	m := newMockLibp2pCommunication(t, []types.NodeID{nodeA, nodeB, nodeC}, logger, 10000)

	a := m.getTransport(nodeA)
	err := a.Start()
	require.NoError(t, err)
	defer a.Stop()

	b := m.getTransport(nodeB)
	err = b.Start()
	require.NoError(t, err)
	defer b.Stop()

	c := m.getTransport(nodeC)
	err = c.Start()
	require.NoError(t, err)
	defer c.Stop()

	t.Log(">>> connecting nodes")

	a.Connect(ctx, m.Nodes())
	b.Connect(ctx, m.Nodes())
	c.Connect(ctx, m.Nodes())

	t.Log(">>> sending messages")
	err = m.sendEventually(ctx, nodeA, "unknownNode", &messagepb.Message{})
	require.Error(t, err)

	nodeBEventsChan := b.EventsOut()
	nodeCEventsChan := c.EventsOut()

	err = a.Send(ctx, nodeB, &messagepb.Message{})
	require.NoError(t, err)
	nodeBEvents := <-nodeBEventsChan
	msg, valid := nodeBEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())

	err = a.Send(ctx, nodeC, &messagepb.Message{})
	require.NoError(t, err)
	nodeCEvents := <-nodeCEventsChan
	msg, valid = nodeCEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())

	t.Log("disconnecting nodes")
	m.disconnect(nodeA, nodeB)
	require.Equal(t, 1, len(b.host.Network().Peers()))

	t.Log(">>> sending messages after disconnection")
	err = m.sendEventually(ctx, nodeA, nodeB, &messagepb.Message{})
	require.NoError(t, err)

	err = a.Send(ctx, nodeC, &messagepb.Message{})
	require.NoError(t, err)
	nodeCEvents = <-nodeCEventsChan
	msg, valid = nodeCEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())

	err = a.Send(ctx, nodeB, &messagepb.Message{})
	require.NoError(t, err)
	nodeBEvents = <-nodeBEventsChan
	msg, valid = nodeBEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())
}
