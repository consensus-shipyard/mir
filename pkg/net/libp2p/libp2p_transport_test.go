package libp2p

import (
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
	libp2putil "github.com/filecoin-project/mir/pkg/util/libp2p"
)

type mockLibp2pCommunication struct {
	t          *testing.T
	params     Params
	membership map[types.NodeID]types.NodeAddress
	hosts      map[types.NodeID]host.Host
	logger     logging.Logger
	transports map[types.NodeID]*Transport
}

func newMockLibp2pCommunication(
	t *testing.T,
	params Params,
	nodeIDs []types.NodeID,
	logger logging.Logger,
) *mockLibp2pCommunication {
	lt := &mockLibp2pCommunication{
		t:          t,
		params:     params,
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
		lt.transports[id], err = NewTransport(params, lt.hosts[id], id, logger)
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

func (m *mockLibp2pCommunication) StopAll() {
	for _, v := range m.transports {
		v.Stop()
	}
}

func (m *mockLibp2pCommunication) CloseHost(transports ...*Transport) {
	for _, v := range transports {
		err := v.host.Close()
		require.NoError(m.t, err)
	}
}

func (m *mockLibp2pCommunication) CloseHostAll() {
	for _, v := range m.transports {
		err := v.host.Close()
		require.NoError(m.t, err)
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

func (m *mockLibp2pCommunication) testEventuallySentMsg(srcNode, dstNode types.NodeID, msg *messagepb.Message) {
	src := m.getTransport(srcNode)

	require.Eventually(m.t,
		func() bool {
			err := src.Send(dstNode, msg)
			if err != nil {
				m.t.Log(err)
			}
			return err == nil
		},
		5*time.Second, 300*time.Millisecond)
}

func (m *mockLibp2pCommunication) testNeverConnected(nodeID1, nodeID2 types.NodeID) {
	src := m.getTransport(nodeID1)
	dst := m.getTransport(nodeID2)

	require.Never(m.t,
		func() bool {
			return network.Connected == src.host.Network().Connectedness(dst.host.ID())
		},
		5*time.Second, 300*time.Millisecond)
}

func (m *mockLibp2pCommunication) testEventuallyNoStreamsBetween(nodeID1, nodeID2 types.NodeID) {
	src := m.getTransport(nodeID1)
	dst := m.getTransport(nodeID2)

	require.Eventually(m.t,
		func() bool {
			n := 0
			conns := src.host.Network().Conns()
			for _, conn := range conns {
				if conn.RemotePeer() == dst.host.ID() {
					for _, s := range conn.GetStreams() {
						if s.Protocol() == m.params.ProtocolID {
							n++
						}
					}
				}

			}
			return n == 0
		},
		10*time.Second, 300*time.Millisecond)
}

func (m *mockLibp2pCommunication) testEventuallyNoStreams(nodeID types.NodeID) {
	v := m.getTransport(nodeID)

	require.Eventually(m.t,
		func() bool {
			return m.getNumberOfStreams(v) == 0
		},
		10*time.Second, 300*time.Millisecond)
}

func (m *mockLibp2pCommunication) testThatSenderIs(events *events.EventList, nodeID types.NodeID) {
	msg, valid := events.Iterator().Next().Type.(*eventpb.Event_MessageReceived)
	require.Equal(m.t, true, valid)
	require.Equal(m.t, msg.MessageReceived.From, nodeID.Pb())

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

func (m *mockLibp2pCommunication) testConnsEmpty() {
	for _, v := range m.transports {
		require.Equal(m.t, 0, len(v.conns))
	}
}

func (m *mockLibp2pCommunication) testNoConnections() {
	for _, v := range m.transports {
		require.Equal(m.t, 0, len(v.host.Network().Conns()))
	}
}

func (m *mockLibp2pCommunication) testNeverMoreThanOneConnectionInProgress(nodeID types.NodeID) {
	src := m.getTransport(nodeID)

	require.Never(m.t,
		func() bool {
			return len(src.conns) > 1
		},
		5*time.Second, 300*time.Millisecond)
}

func (m *mockLibp2pCommunication) getNumberOfStreams(v *Transport) int {
	n := 0
	conns := v.host.Network().Conns()
	for _, c := range conns {
		for _, s := range c.GetStreams() {
			if s.Protocol() == m.params.ProtocolID {
				n++
			}
		}
	}
	return n
}

func TestLibp2p_Sending(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")
	nodeE := types.NodeID("e")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAll()

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

	a.Connect(initialNodes)
	b.Connect(initialNodes)
	c.Connect(initialNodes)
	d.Connect(initialNodes)

	m.testEventuallyConnected(nodeA, nodeB)
	m.testEventuallyConnected(nodeA, nodeC)
	m.testEventuallyConnected(nodeA, nodeD)
	m.testEventuallyConnected(nodeB, nodeC)
	m.testEventuallyConnected(nodeB, nodeD)
	m.testEventuallyConnected(nodeC, nodeD)

	t.Log(">>> sending messages")
	err = a.Send("unknownNode", &messagepb.Message{})
	require.ErrorIs(t, err, ErrUnknownNode)

	err = a.Send(nodeE, &messagepb.Message{})
	require.ErrorAs(t, err, &ErrUnknownNode)

	nodeBEventsChan := b.EventsOut()
	nodeCEventsChan := c.EventsOut()

	err = a.Send(nodeB, &messagepb.Message{})
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	err = a.Send(nodeC, &messagepb.Message{})
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeCEventsChan, nodeA)

	t.Log(">>> disconnecting nodes")
	m.disconnect(nodeA, nodeB)
	m.testNeverConnected(nodeA, nodeB)

	t.Log(">>> sending messages after disconnection")
	m.testEventuallySentMsg(nodeA, nodeB, &messagepb.Message{})
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)
	m.testEventuallySentMsg(nodeA, nodeC, &messagepb.Message{})
	m.testThatSenderIs(<-nodeCEventsChan, nodeA)

	t.Log(">>> cleaning")
	m.StopAll()
	m.testEventuallyNoStreams(nodeA)
	m.testConnsEmpty()

	m.CloseHostAll()
	m.testNoConnections()
}

func TestLibp2p_Connecting(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAll()

	t.Log(">>> connecting nodes")
	t.Log("membership")
	t.Log(m.membership)

	initialNodes := m.Membership(nodeA, nodeB, nodeC)

	a.Connect(initialNodes)
	b.Connect(initialNodes)
	c.Connect(initialNodes)

	m.testEventuallyConnected(nodeA, nodeB)
	m.testEventuallyConnected(nodeA, nodeC)
	m.testEventuallyConnected(nodeB, nodeC)
	m.testNeverConnected(nodeA, nodeD)
	m.testNeverConnected(nodeB, nodeD)
	m.testNeverConnected(nodeC, nodeD)

	t.Log(">>> reconfigure nodes")
	newNodes := m.Membership(nodeA, nodeB, nodeD)
	t.Log("new membership")
	t.Log(newNodes)

	a.Connect(newNodes)
	a.CloseOldConnections(newNodes)

	b.Connect(newNodes)
	b.CloseOldConnections(newNodes)

	d.Connect(newNodes)
	d.CloseOldConnections(newNodes)

	m.testEventuallyConnected(nodeA, nodeB)
	m.testEventuallyConnected(nodeA, nodeD)
	m.testEventuallyConnected(nodeB, nodeD)

	m.testEventuallyNoStreamsBetween(nodeA, nodeC)
	m.testEventuallyNoStreamsBetween(nodeB, nodeC)
	m.testEventuallyNoStreamsBetween(nodeD, nodeC)

	t.Log(">>> cleaning")
	m.StopAll()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testConnsEmpty()

	m.CloseHostAll()
	m.testNoConnections()
}

func TestLibp2p_SendingWithTwoNodes(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAll()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)

	t.Log(">>> connecting nodes")

	initialNodes := m.Membership(nodeA, nodeB)

	t.Log("membership")
	t.Log(initialNodes)

	a.Connect(initialNodes)
	b.Connect(initialNodes)

	m.testEventuallyConnected(nodeA, nodeB)
	m.testEventuallyConnected(nodeB, nodeA)

	t.Log(">>> sending messages")

	nodeBEventsChan := b.EventsOut()

	m.testEventuallySentMsg(nodeA, nodeB, &messagepb.Message{})
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> disconnecting nodes")
	m.disconnect(nodeA, nodeB)
	m.testNeverConnected(nodeA, nodeB)

	t.Log(">>> sending messages after disconnection")
	m.testEventuallySentMsg(nodeA, nodeB, &messagepb.Message{})

	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> cleaning")
	m.StopAll()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()

	m.CloseHostAll()
	m.testNoConnections()
}

func TestLibp2p_SendingWithTwoNodesSyncMode(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAll()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)

	t.Log(">>> connecting nodes")

	initialNodes := m.Membership(nodeA, nodeB)

	t.Log("membership")
	t.Log(initialNodes)

	a.Connect(initialNodes)
	b.Connect(initialNodes)

	m.testEventuallyConnected(nodeA, nodeB)

	t.Log(">>> sending messages")

	nodeBEventsChan := b.EventsOut()

	a.WaitFor(2)
	err := a.Send(nodeB, &messagepb.Message{})
	require.NoError(t, err)

	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> disconnecting nodes")
	m.disconnect(nodeA, nodeB)
	m.testNeverConnected(nodeA, nodeB)

	t.Log(">>> sending messages after disconnection")
	m.testEventuallySentMsg(nodeA, nodeB, &messagepb.Message{})

	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> cleaning")
	m.StopAll()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()

	m.CloseHostAll()
	m.testNoConnections()
}

func TestLibp2p_Sending2NodesNonBlock(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAll()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)

	t.Log(">>> connecting nodes")

	initialNodes := m.Membership(nodeA, nodeB)

	t.Log("membership")
	t.Log(initialNodes)

	a.Connect(initialNodes)
	b.Connect(initialNodes)
	m.testEventuallyConnected(nodeA, nodeB)
	a.WaitFor(2)

	t.Log(">>> send a message")
	err := a.Send(nodeB, &messagepb.Message{})
	require.NoError(t, err)

	nodeBEventsChan := b.incomingMessages

	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> blocking message")

	go func() {
		b.incomingMessages <- events.ListOf(
			events.MessageReceived(types.ModuleID("1"), "blocker", &messagepb.Message{}),
		)
		b.incomingMessages <- events.ListOf(
			events.MessageReceived(types.ModuleID("1"), "blocker", &messagepb.Message{}),
		)
	}()
	err = a.Send(nodeB, &messagepb.Message{})
	time.Sleep(5 * time.Second)
	require.NoError(t, err)

	t.Log(">>> cleaning")
	m.StopAll()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()

	m.CloseHostAll()
	m.testNoConnections()

	m.testThatSenderIs(<-nodeBEventsChan, "blocker")
	m.testThatSenderIs(<-nodeBEventsChan, "blocker")
}

// Test we don't get two elements in the node table corresponding to the same node.
func TestLibp2p_OneConnectionInProgress(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeE := types.NodeID("e")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA}, logger)

	a := m.transports[nodeA]
	m.StartAll()
	defer m.StopAll()

	eAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/0")
	require.NoError(t, err)

	t.Log(">>> connecting nodes")

	initialNodes := m.Membership(nodeA)
	initialNodes[nodeE] = eAddr

	a.Connect(initialNodes)
	a.Send(nodeE, &messagepb.Message{}) // nolint

	m.testNeverMoreThanOneConnectionInProgress(nodeA)
}

// Test that when a node fails to connect first time it will be trying to connect.
func TestLibp2p_OpeningConnectionAfterFail(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	b := m.transports[nodeB]

	bBadAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/100/p2p/%s", b.host.ID()))
	require.NoError(t, err)

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)

	initialNodes := m.Membership(nodeA)
	initialNodes[nodeB] = bBadAddr
	t.Log(initialNodes)

	t.Log(">>> connecting to a failed node")
	err = a.Start()
	require.NoError(t, err)
	a.params.MaxRetries = 1
	a.params.MaxConnectingTimeout = 1 * time.Second
	a.params.MaxRetryTimeout = 0
	a.Connect(initialNodes)
	m.testNeverConnected(nodeA, nodeB)

	t.Log(">>> connecting to the restarted node")
	err = b.Start()
	require.NoError(t, err)
	initialNodes = m.Membership(nodeA, nodeB)
	require.NoError(t, err)
	a.Connect(initialNodes)
	m.testEventuallyConnected(nodeA, nodeB)

	t.Log(">>> cleaning")
	m.StopAll()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()

	m.CloseHostAll()
	m.testNoConnections()
}

func TestLibp2p_TwoNodesBasic(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAll()

	t.Log(">>> connecting nodes")
	t.Log("membership")
	t.Log(m.membership)

	initialNodes := m.Membership(nodeA, nodeB)

	a.Connect(initialNodes)
	b.Connect(initialNodes)
	m.testEventuallyConnected(nodeA, nodeB)

	t.Log(">>> cleaning")
	m.StopAll()
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()
	m.CloseHostAll()
	m.testNoConnections()
}
