package libp2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
	libp2putil "github.com/filecoin-project/mir/pkg/util/libp2p"
)

type mockLibp2pCommunication struct {
	t          *testing.T
	membership map[types.NodeID]types.NodeAddress
	hosts      map[types.NodeID]host.Host
	logger     logging.Logger
	transports map[types.NodeID]*Transport
}

func newMockLibp2pCommunication(t *testing.T, nodeIDs []types.NodeID, logger logging.Logger) *mockLibp2pCommunication {
	lt := &mockLibp2pCommunication{
		t:          t,
		membership: make(map[types.NodeID]types.NodeAddress, len(nodeIDs)),
		hosts:      make(map[types.NodeID]host.Host),
		transports: make(map[types.NodeID]*Transport),
		logger:     logger,
	}

	var err error
	for i, id := range nodeIDs {
		hostAddr := libp2putil.NewFreeHostAddr()
		lt.hosts[id] = libp2putil.NewDummyHost(i, hostAddr)
		lt.membership[id] = types.NodeAddress(libp2putil.NewDummyMultiaddr(i, hostAddr))
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

func (m *mockLibp2pCommunication) Stop(transports ...*Transport) {
	for _, v := range transports {
		v.Stop()
	}
}

func (m *mockLibp2pCommunication) StopAll(transports ...*Transport) {
	for _, v := range m.transports {
		v.Stop()
	}
}

func (m *mockLibp2pCommunication) Start(transports ...*Transport) {
	for _, v := range transports {
		err := v.Start()
		require.NoError(m.t, err, "starting node", v.ownID)
	}
}

func (m *mockLibp2pCommunication) StartAll() {
	for _, v := range m.transports {
		err := v.Start()
		require.NoError(m.t, err, "starting node", v.ownID)
	}
}

func (m *mockLibp2pCommunication) Membership(ids ...types.NodeID) map[types.NodeID]types.NodeAddress {
	nodeMap := make(map[types.NodeID]types.NodeAddress)
	for _, id := range ids {
		addr, ok := m.membership[id]
		if !ok {
			m.t.Fatalf("failed to get addr for a new node %v", id)
		}
		nodeMap[id] = addr
	}
	return nodeMap
}

func (m *mockLibp2pCommunication) FourTransports(nodeID ...types.NodeID) (*Transport, *Transport, *Transport, *Transport) {
	if len(nodeID) != 4 {
		m.t.Fatalf("want 4 node IDs, have %d", len(nodeID))
	}

	ts := [4]*Transport{}

	for i, v := range nodeID {
		ts[i] = m.transports[v]
	}

	return ts[0], ts[1], ts[2], ts[3]
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

func (m *mockLibp2pCommunication) testConnectedToNodesEventually(node types.NodeID, nodes ...types.NodeID) error {
	src := m.getTransport(node)

	timeout := time.After(10 * time.Second)
	checkTicker := time.NewTicker(1000 * time.Millisecond)
	defer checkTicker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("%v is not connected to %v", node, nodes)
		case <-checkTicker.C:
			connections := 0
			m.t.Logf("%v checks connections to %v\n", node, nodes)
			for i := range nodes {
				dst := m.transports[nodes[i]]
				m.t.Logf("%v/%v connection: %v\n", node, nodes[i], src.host.Network().ConnsToPeer(dst.host.ID()))
				if len(src.host.Network().ConnsToPeer(dst.host.ID())) != 0 {
					connections++
				}
			}
			if connections == len(nodes) {
				m.t.Logf("%v connected to %v\n", node, nodes)
				return nil
			}
		}
	}
}

func (m *mockLibp2pCommunication) testNoConnectionBetweenNodesEventually(nodeID1, nodeID2 types.NodeID) error {
	n1 := m.getTransport(nodeID1)
	n2 := m.getTransport(nodeID2)

	timeout := time.After(10 * time.Second)
	checkTicker := time.NewTicker(1000 * time.Millisecond)
	defer checkTicker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("%v has connection to %v", nodeID1, nodeID2)
		case <-checkTicker.C:
			m.t.Logf("%v checks connections to %v\n", nodeID1, nodeID2)
			if len(n1.host.Network().ConnsToPeer(n2.host.ID())) == 0 {
				m.t.Logf("%v not connected to %v\n", nodeID1, nodeID2)
				return nil
			}
		}
	}
}

func TestLibp2p_Sending(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")

	m := newMockLibp2pCommunication(t, []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAll()
	defer m.StopAll()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)
	require.Equal(t, nodeC, c.ownID)

	t.Log(">>> connecting nodes")

	initialNodes := m.Membership(nodeA, nodeB, nodeC, nodeD)

	a.Connect(ctx, initialNodes)
	b.Connect(ctx, initialNodes)
	c.Connect(ctx, initialNodes)
	d.Connect(ctx, initialNodes)

	t.Log(">>> sending messages")
	err := m.sendEventually(ctx, nodeA, "unknownNode", &messagepb.Message{})
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

	t.Log(">>> disconnecting nodes")
	m.disconnect(nodeA, nodeB)
	require.Equal(t, 2, len(b.host.Network().Peers()))

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

func TestLibp2p_Connecting(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")

	m := newMockLibp2pCommunication(t, []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAll()
	defer m.StopAll()

	t.Log(">>> connecting nodes")

	initialNodes := m.Membership(nodeA, nodeB, nodeC)

	a.Connect(ctx, initialNodes)
	b.Connect(ctx, initialNodes)
	c.Connect(ctx, initialNodes)

	require.NoError(t, m.testConnectedToNodesEventually(nodeA, nodeB, nodeC))
	require.NoError(t, m.testNoConnectionBetweenNodesEventually(nodeA, nodeD))

	t.Log(">>> reconfigure nodes")

	newNodes := m.Membership(nodeA, nodeB, nodeD)

	a.Connect(ctx, newNodes)
	a.CloseOldConnections(newNodes)

	b.Connect(ctx, newNodes)
	b.CloseOldConnections(newNodes)

	d.Connect(ctx, newNodes)
	d.CloseOldConnections(newNodes)

	require.NoError(t, m.testConnectedToNodesEventually(nodeA, nodeB, nodeD))
	require.NoError(t, m.testConnectedToNodesEventually(nodeB, nodeA, nodeD))

	require.NoError(t, m.testNoConnectionBetweenNodesEventually(nodeA, nodeC))
	require.NoError(t, m.testNoConnectionBetweenNodesEventually(nodeB, nodeC))
	require.NoError(t, m.testNoConnectionBetweenNodesEventually(nodeD, nodeC))
}
