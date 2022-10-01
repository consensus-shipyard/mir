package libp2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
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

func (m *mockLibp2pCommunication) testEventuallySentMsg(ctx context.Context, srcNode, dstNode types.NodeID, msg *messagepb.Message) {
	src := m.getTransport(srcNode)

	require.Eventually(m.t,
		func() bool {
			err := src.Send(ctx, dstNode, msg)
			return err == nil
		},
		5*time.Second, 300*time.Millisecond)
}

func (m *mockLibp2pCommunication) testEventuallyNotConnected(nodeID1, nodeID2 types.NodeID) {
	src := m.getTransport(nodeID1)
	dst := m.getTransport(nodeID2)

	require.Eventually(m.t,
		func() bool {
			fmt.Println(222, src.host.Network().ConnsToPeer(dst.host.ID()))
			return network.NotConnected == src.host.Network().Connectedness(dst.host.ID())
		},
		5*time.Second, 300*time.Millisecond)
}

func (m *mockLibp2pCommunication) testEventuallyConnected(nodeID1, nodeID2 types.NodeID) {
	src := m.getTransport(nodeID1)
	dst := m.getTransport(nodeID2)

	require.Eventually(m.t,
		func() bool {
			return network.Connected == src.host.Network().Connectedness(dst.host.ID())
		},
		10*time.Second, 100*time.Millisecond)
}

func TestLibp2p_Sending(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")
	nodeE := types.NodeID("e")

	m := newMockLibp2pCommunication(t, []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAll()
	defer m.StopAll()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)
	require.Equal(t, nodeC, c.ownID)
	require.Equal(t, nodeD, d.ownID)

	eAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/0")
	require.NoError(t, err)

	t.Log(">>> connecting nodes")

	initialNodes := m.Membership(nodeA, nodeB, nodeC, nodeD)
	initialNodes[nodeE] = eAddr

	t.Log("membership")
	t.Log(initialNodes)

	a.Connect(ctx, initialNodes)
	b.Connect(ctx, initialNodes)
	c.Connect(ctx, initialNodes)
	d.Connect(ctx, initialNodes)

	m.testEventuallyConnected(nodeA, nodeB)
	m.testEventuallyConnected(nodeA, nodeC)
	m.testEventuallyConnected(nodeA, nodeD)
	m.testEventuallyConnected(nodeB, nodeC)
	m.testEventuallyConnected(nodeB, nodeD)
	m.testEventuallyConnected(nodeC, nodeD)

	t.Log(">>> sending messages")
	var nodeErr *nodeInfoError
	err = a.Send(ctx, "unknownNode", &messagepb.Message{})
	require.ErrorAs(t, err, &nodeErr)

	var streamErr *nilStreamError
	err = a.Send(ctx, nodeE, &messagepb.Message{})
	require.ErrorAs(t, err, &streamErr)

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
	m.testEventuallyNotConnected(nodeA, nodeB)
	require.Equal(t, 2, len(b.host.Network().Peers()))

	t.Log(">>> sending messages after disconnection")
	m.testEventuallySentMsg(ctx, nodeA, nodeB, &messagepb.Message{})

	nodeBEvents = <-nodeBEventsChan
	msg, valid = nodeBEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(t, true, valid)
	require.Equal(t, msg.MessageReceived.From, nodeA.Pb())

	m.testEventuallySentMsg(ctx, nodeA, nodeC, &messagepb.Message{})
	nodeCEvents = <-nodeCEventsChan
	msg, valid = nodeCEvents.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
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
	t.Log("membership")
	t.Log(m.membership)

	initialNodes := m.Membership(nodeA, nodeB, nodeC)

	a.Connect(ctx, initialNodes)
	b.Connect(ctx, initialNodes)
	c.Connect(ctx, initialNodes)

	m.testEventuallyConnected(nodeA, nodeB)
	m.testEventuallyConnected(nodeA, nodeC)
	m.testEventuallyConnected(nodeB, nodeC)
	m.testEventuallyNotConnected(nodeA, nodeD)
	m.testEventuallyNotConnected(nodeB, nodeD)
	m.testEventuallyNotConnected(nodeC, nodeD)

	t.Log(">>> reconfigure nodes")
	newNodes := m.Membership(nodeA, nodeB, nodeD)
	t.Log("new membership")
	t.Log(newNodes)

	a.Connect(ctx, newNodes)
	a.CloseOldConnections(newNodes)

	b.Connect(ctx, newNodes)
	b.CloseOldConnections(newNodes)

	d.Connect(ctx, newNodes)
	d.CloseOldConnections(newNodes)

	m.testEventuallyConnected(nodeA, nodeB)
	m.testEventuallyConnected(nodeA, nodeD)
	m.testEventuallyConnected(nodeB, nodeD)

	m.testEventuallyNotConnected(nodeA, nodeC)
	m.testEventuallyNotConnected(nodeB, nodeC)
	m.testEventuallyNotConnected(nodeD, nodeC)
}
