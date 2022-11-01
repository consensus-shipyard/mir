package libp2p

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

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
		m.t.Fatalf("failed to get transport for node: %v", sourceID)
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

func (m *mockLibp2pCommunication) StopTransports(transports ...*Transport) {
	for _, v := range transports {
		v.Stop()
	}
}

func (m *mockLibp2pCommunication) StopAllTransports() {
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

func (m *mockLibp2pCommunication) CloseAllHosts() {
	for _, v := range m.transports {
		err := v.host.Close()
		require.NoError(m.t, err)
	}
}

func (m *mockLibp2pCommunication) StartTransports(transports ...*Transport) {
	for _, v := range transports {
		err := v.Start()
		require.NoError(m.t, err, "starting node", v.ownID)
	}
}

func (m *mockLibp2pCommunication) StartAllTransports() {
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

func (m *mockLibp2pCommunication) testEventuallyNotConnected(nodeID1, nodeID2 types.NodeID) {
	n1 := m.getTransport(nodeID1)
	n2 := m.getTransport(nodeID2)

	require.Eventually(m.t,
		func() bool {
			return network.NotConnected == n1.host.Network().Connectedness(n2.host.ID()) &&
				!m.streamExist(n1, n2) && !m.streamExist(n2, n1)
		},
		10*time.Second, 300*time.Millisecond)
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

// testEventuallyHasConnection tests that there is a connection between the nodes initiated by initiator.
//
// The nodes are connected if the conditions are met:
// - there is a connection between initiator and responder nodes
// - there is a stream from initiator to responder
// - the stream is processed by the target protocol
func (m *mockLibp2pCommunication) testEventuallyHasConnection(initiator, responder types.NodeID) {
	n1 := m.getTransport(initiator)
	n2 := m.getTransport(responder)

	require.Eventually(m.t,
		func() bool {
			return network.Connected == n1.host.Network().Connectedness(n2.host.ID()) &&
				m.streamExist(n1, n2)
		},
		10*time.Second, 100*time.Millisecond)
}

// testEventuallyConnected tests that there is a bidirectional connection between the nodes.
//
// The nodes are connected if the conditions are met:
// - there is a connection between nodeID1 and nodeID2 nodes
// - there is a stream from nodeID1 to nodeID2
// - there is a stream from nodeID2 to nodeID1
// - the streams are processed by the target protocol
func (m *mockLibp2pCommunication) testEventuallyConnected(nodeID1, nodeID2 types.NodeID) {
	n1 := m.getTransport(nodeID1)
	n2 := m.getTransport(nodeID2)

	require.Eventually(m.t,
		func() bool {
			return network.Connected == n1.host.Network().Connectedness(n2.host.ID()) &&
				m.streamExist(n1, n2) && m.streamExist(n2, n1)
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

func (m *mockLibp2pCommunication) streamExist(src, dst *Transport) bool {
	conns := src.host.Network().Conns()

	for _, c := range conns {
		for _, s := range c.GetStreams() {
			if s.Protocol() == m.params.ProtocolID {
				if s.Conn().RemotePeer() == dst.host.ID() {
					return true
				}
			}
		}
	}
	return false
}

func TestLibp2p_Sending(t *testing.T) {
	defer goleak.VerifyNone(t)

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")
	nodeE := types.NodeID("e")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAllTransports()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)
	require.Equal(t, nodeC, c.ownID)
	require.Equal(t, nodeD, d.ownID)

	eAddr := libp2putil.NewFakeMultiaddr(100, 0)

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
	err := a.Send("unknownNode", &messagepb.Message{})
	require.ErrorIs(t, err, ErrUnknownNode)

	err = a.Send(nodeE, &messagepb.Message{})
	require.ErrorIs(t, err, ErrNilStream)

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
	m.testEventuallyNotConnected(nodeA, nodeB)

	t.Log(">>> sending messages after disconnection")
	m.testEventuallySentMsg(nodeA, nodeB, &messagepb.Message{})
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)
	m.testEventuallySentMsg(nodeA, nodeC, &messagepb.Message{})
	m.testThatSenderIs(<-nodeCEventsChan, nodeA)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreams(nodeA)
	m.testConnsEmpty()

	m.CloseAllHosts()
	m.testNoConnections()
}

func TestLibp2p_Connecting(t *testing.T) {
	defer goleak.VerifyNone(t)

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAllTransports()

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
	m.testEventuallyNotConnected(nodeA, nodeD)
	m.testEventuallyNotConnected(nodeB, nodeD)
	m.testEventuallyNotConnected(nodeC, nodeD)

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
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testConnsEmpty()

	m.CloseAllHosts()
	m.testNoConnections()
}

func TestLibp2p_SendingWithTwoNodes(t *testing.T) {
	defer goleak.VerifyNone(t)

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAllTransports()

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
	m.testEventuallyNotConnected(nodeA, nodeB)

	t.Log(">>> sending messages after disconnection")
	m.testEventuallySentMsg(nodeA, nodeB, &messagepb.Message{})

	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()

	m.CloseAllHosts()
	m.testNoConnections()
}

func TestLibp2p_Messaging(t *testing.T) {
	defer goleak.VerifyNone(t)

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	require.Equal(t, nodeA, a.ownID)
	b := m.transports[nodeB]
	require.Equal(t, nodeB, b.ownID)
	m.StartAllTransports()

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

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		<-ctx.Done()

		t.Log(">>> cleaning")
		m.StopAllTransports()
		m.testEventuallyNoStreamsBetween(nodeA, nodeB)
		m.testEventuallyNoStreams(nodeA)
		m.testEventuallyNoStreams(nodeB)
		m.testConnsEmpty()

		m.CloseAllHosts()
		m.testNoConnections()
	}(ctx)

	sent := 0
	sentBeforeDisconnect := 0
	received := 0
	disconnect := make(chan struct{})

	testTimeDuration := time.Duration(15)
	testTimer := time.NewTimer(testTimeDuration * time.Second)
	disconnectTimer := time.NewTimer(testTimeDuration / 3 * time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-disconnect
		t.Log(">>> disconnecting")
		m.disconnect(nodeA, nodeB)
	}()

	send := time.NewTicker(300 * time.Millisecond)
	defer send.Stop()

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-disconnectTimer.C:
				t.Log("disconnect nodes")
				sentBeforeDisconnect = sent
				disconnect <- struct{}{}
			case <-send.C:
				err := a.Send(nodeB, &messagepb.Message{})
				if err != nil {
					m.t.Log(err)
				} else {
					sent++
				}
			}
		}
	}(ctx)

	for {
		select {
		case msg := <-nodeBEventsChan:
			m.testThatSenderIs(msg, nodeA)
			received++
		case <-testTimer.C:
			cancel()
			wg.Wait()
			if received <= sentBeforeDisconnect+(sent-sentBeforeDisconnect)/2 {
				t.Fail()
			}
			t.Log("sent before disconnect: ", sentBeforeDisconnect)
			t.Log("sent: ", sent)
			t.Log("received: ", received)
			return
		}
	}
}

func TestLibp2p_SendingWithTwoNodesInSyncMode(t *testing.T) {
	defer goleak.VerifyNone(t)

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)
	a := m.transports[nodeA]
	b := m.transports[nodeB]
	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)
	m.StartAllTransports()
	initialNodes := m.Membership(nodeA, nodeB)
	t.Log("membership")
	t.Log(initialNodes)

	t.Log(">>> connecting nodes")
	a.Connect(initialNodes)
	b.Connect(initialNodes)
	a.WaitFor(2)
	m.testEventuallyConnected(nodeA, nodeB)

	t.Log(">>> sending messages")
	nodeBEventsChan := b.EventsOut()
	err := a.Send(nodeB, &messagepb.Message{})
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> disconnecting nodes")
	m.disconnect(nodeA, nodeB)
	m.testEventuallyNotConnected(nodeA, nodeB)

	t.Log(">>> sending messages after disconnection")
	m.testEventuallySentMsg(nodeA, nodeB, &messagepb.Message{})
	m.testEventuallyConnected(nodeA, nodeB)

	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()

	m.CloseAllHosts()
	m.testNoConnections()
}

func TestLibp2p_SendingWithTwoNodesNonBlock(t *testing.T) {
	defer goleak.VerifyNone(t)

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAllTransports()

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
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()

	m.CloseAllHosts()
	m.testNoConnections()

	m.testThatSenderIs(<-nodeBEventsChan, "blocker")
	m.testThatSenderIs(<-nodeBEventsChan, "blocker")
}

// Test we don't get two elements in the node table corresponding to the same node.
func TestLibp2p_OneConnectionInProgress(t *testing.T) {
	defer goleak.VerifyNone(t)

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeE := types.NodeID("e")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA}, logger)

	a := m.transports[nodeA]
	m.StartAllTransports()

	eAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/0")
	require.NoError(t, err)

	t.Log(">>> connecting nodes")

	initialNodes := m.Membership(nodeA)
	initialNodes[nodeE] = eAddr

	t.Log(">>> send a message")
	a.Connect(initialNodes)
	a.Send(nodeE, &messagepb.Message{}) // nolint
	m.testNeverMoreThanOneConnectionInProgress(nodeA)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.CloseAllHosts()
	m.testNoConnections()
}

// Test that when a node fails to connect first time it will be trying to connect.
func TestLibp2p_OpeningConnectionAfterFail(t *testing.T) {
	defer goleak.VerifyNone(t)

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
	m.testEventuallyHasConnection(nodeA, nodeB)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()

	m.CloseAllHosts()
	m.testNoConnections()
}

func TestLibp2p_TwoNodesBasic(t *testing.T) {
	defer goleak.VerifyNone(t)

	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAllTransports()

	t.Log(">>> connecting nodes")
	t.Log("membership")
	t.Log(m.membership)

	initialNodes := m.Membership(nodeA, nodeB)

	a.Connect(initialNodes)
	b.Connect(initialNodes)
	m.testEventuallyConnected(nodeA, nodeB)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnsEmpty()
	m.CloseAllHosts()
	m.testNoConnections()
}
