package libp2p

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	checkpointpbmsgs "github.com/filecoin-project/mir/pkg/pb/checkpointpb/msgs"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/transportpb"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
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

	for i, id := range nodeIDs {
		hostAddr := libp2putil.NewFreeHostAddr()
		lt.hosts[id] = libp2putil.NewDummyHost(i, hostAddr)
		lt.membership[id] = types.NodeAddress(libp2putil.NewDummyMultiaddr(i, hostAddr))
		lt.transports[id] = NewTransport(params, id, lt.hosts[id], logger)
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
	m.t.Log(">>> stop transports")
	for _, v := range transports {
		v.Stop()
	}
}

func (m *mockLibp2pCommunication) StopAllTransports() {
	m.t.Log(">>> stop transports")
	for _, v := range m.transports {
		v.Stop()
	}
}

func (m *mockLibp2pCommunication) CloseHosts(transports ...*Transport) {
	m.t.Log(">>> close hosts")
	for _, v := range transports {
		err := v.host.Close()
		require.NoError(m.t, err)
	}
}

func (m *mockLibp2pCommunication) CloseAllHosts() {
	m.t.Log(">>> close hosts")
	for _, v := range m.transports {
		err := v.host.Close()
		if err != nil {
			m.t.Fatalf("failed close host %v: %v", v.ownID, err)
		}
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

func (m *mockLibp2pCommunication) testEventuallyNoStreams(nodeIDs ...types.NodeID) {
	for _, nodeID := range nodeIDs {
		v := m.getTransport(nodeID)

		require.Eventually(m.t,
			func() bool {
				m.printStreams(nodeID)
				return m.getNumberOfStreams(v) == 0
			},
			10*time.Second, 300*time.Millisecond)
	}
}

func (m *mockLibp2pCommunication) testThatSenderIs(events *events.EventList, nodeID types.NodeID) {
	tEvent, valid := events.Iterator().Next().Type.(*eventpb.Event_Transport)
	require.True(m.t, valid)
	msg, valid := tEvent.Transport.Type.(*transportpb.Event_MessageReceived)
	require.True(m.t, valid)
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
		15*time.Second, 100*time.Millisecond)
}

func (m *mockLibp2pCommunication) testConnectionsEmpty() {
	for _, v := range m.transports {
		require.Equal(m.t, 0, len(v.connections))
	}
}

func (m *mockLibp2pCommunication) testNoHostConnections() {
	for _, v := range m.transports {
		require.Equal(m.t, 0, len(v.host.Network().Conns()))
	}
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

func (m *mockLibp2pCommunication) printStreams(node types.NodeID) { // nolint
	v := m.getTransport(node)
	conns := v.host.Network().Conns()
	for _, c := range conns {
		for _, s := range c.GetStreams() {
			if s.Protocol() == m.params.ProtocolID {
				fmt.Printf(">>>>> node %v streams: %v\n", node, c)
			}
		}
	}
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

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// TestNewTransport tests that the transport can be created.
func TestNewTransport(t *testing.T) {
	addr := libp2putil.NewFreeHostAddr()
	h := libp2putil.NewDummyHostNoListen(1, addr)
	_ = NewTransport(DefaultParams(), types.NodeID("a"), h, logging.ConsoleDebugLogger)
	err := h.Close()
	require.NoError(t, err)
}

// TestTransportStartStop tests that the transport can be started and stopped.
func TestTransportStartStop(t *testing.T) {
	addr := libp2putil.NewFreeHostAddr()
	h := libp2putil.NewDummyHostNoListen(1, addr)
	tr := NewTransport(DefaultParams(), types.NodeID("a"), h, logging.ConsoleDebugLogger)

	err := tr.Start()
	require.NoError(t, err)
	tr.Stop()

	err = h.Close()
	require.NoError(t, err)
}

// TestConnectTwoNodes that two nodes can be connected.
func TestConnectTwoNodes(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)
	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAllTransports()

	t.Log(">>> membership")
	initialNodes := m.Membership(nodeA, nodeB)
	t.Log(initialNodes)

	t.Log(">>> connecting nodes")
	a.Connect(initialNodes)
	b.Connect(initialNodes)
	m.testEventuallyConnected(nodeA, nodeB)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA, nodeB)
	m.testConnectionsEmpty()
	m.CloseAllHosts()
	m.testNoHostConnections()
}

// TestCallSendWithoutConnect test that method `Send` returns an error when it is called without `Connect`.
func TestCallSendWithoutConnect(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)
	a := m.transports[nodeA]
	m.StartAllTransports()

	err := a.Send(nodeB, &messagepb.Message{})
	require.Error(t, err)

	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA, nodeB)
	m.testConnectionsEmpty()
	m.CloseAllHosts()
	m.testNoHostConnections()
}

// TestSendReceive tests that messages can be sent and received.
func TestSendReceive(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")
	// nodeE := types.NodeID("e")

	testMsg := checkpointpbmsgs.Checkpoint("", 0, 0, []byte{}, []byte{}) // Just a random message, could be anything.

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAllTransports()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)
	require.Equal(t, nodeC, c.ownID)
	require.Equal(t, nodeD, d.ownID)

	// eAddr := libp2putil.NewFakeMultiaddr(100, 0)

	t.Log(">>> membership")
	initialNodes := m.Membership(nodeA, nodeB, nodeC, nodeD)
	// initialNodes[nodeE] = eAddr
	t.Log(initialNodes)

	t.Log(">>> connecting nodes")
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

	nodeBEventsChan := b.EventsOut()
	nodeCEventsChan := c.EventsOut()

	err := a.Send(nodeB, testMsg.Pb())
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	err = a.Send(nodeC, testMsg.Pb())
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeCEventsChan, nodeA)

	t.Log(">>> disconnecting nodes")
	m.disconnect(nodeA, nodeB)
	m.testEventuallyNotConnected(nodeA, nodeB)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testEventuallyNoStreams(nodeC)
	m.testEventuallyNoStreams(nodeD)
	m.testConnectionsEmpty()
	m.CloseAllHosts()
	m.testNoHostConnections()
}

// TestSendReceive tests that even empty messages can be sent and received.
// TODO: Enable this test ASAP (by removing the 'x'), when support for nil values in generated types is implemented.
// Also remove the nolint:unused tag.
func xTestSendReceiveEmptyMessage(t *testing.T) { // nolint:unused
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")
	// nodeE := types.NodeID("e")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAllTransports()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)
	require.Equal(t, nodeC, c.ownID)
	require.Equal(t, nodeD, d.ownID)

	// eAddr := libp2putil.NewFakeMultiaddr(100, 0)

	t.Log(">>> membership")
	initialNodes := m.Membership(nodeA, nodeB, nodeC, nodeD)
	// initialNodes[nodeE] = eAddr
	t.Log(initialNodes)

	t.Log(">>> connecting nodes")
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

	nodeBEventsChan := b.EventsOut()
	nodeCEventsChan := c.EventsOut()

	err := a.Send(nodeB, &messagepb.Message{})
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	err = a.Send(nodeC, &messagepb.Message{})
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeCEventsChan, nodeA)

	t.Log(">>> disconnecting nodes")
	m.disconnect(nodeA, nodeB)
	m.testEventuallyNotConnected(nodeA, nodeB)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testEventuallyNoStreams(nodeC)
	m.testEventuallyNoStreams(nodeD)
	m.testConnectionsEmpty()
	m.CloseAllHosts()
	m.testNoHostConnections()
}

// TestConnect test that connections can be changed.
func TestConnect(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")
	nodeC := types.NodeID("c")
	nodeD := types.NodeID("d")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB, nodeC, nodeD}, logger)

	a, b, c, d := m.FourTransports(nodeA, nodeB, nodeC, nodeD)
	m.StartAllTransports()

	t.Log(">>> membership")
	initialNodes := m.Membership(nodeA, nodeB, nodeC)
	t.Log(initialNodes)

	t.Log(">>> connecting nodes")
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
	m.testConnectionsEmpty()

	m.CloseAllHosts()
	m.testNoHostConnections()
}

func TestMessaging(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	testMsg := checkpointpbmsgs.Checkpoint("", 0, 0, []byte{}, []byte{}) // Just a random message, could be anything.

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	require.Equal(t, nodeA, a.ownID)
	b := m.transports[nodeB]
	require.Equal(t, nodeB, b.ownID)
	m.StartAllTransports()

	t.Log(">>> membership")
	initialNodes := m.Membership(nodeA, nodeB)
	t.Log(initialNodes)

	t.Log(">>> connecting nodes")
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
		m.testConnectionsEmpty()

		m.CloseAllHosts()
		m.testNoHostConnections()
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
				err := a.Send(nodeB, testMsg.Pb())
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

// TestSendingReceiveWithWaitFor tests that the transport operates normally if WaitFor call is used.
func TestSendingReceiveWithWaitFor(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	testMsg := checkpointpbmsgs.Checkpoint("", 0, 0, []byte{}, []byte{}) // Just a random message, could be anything.

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)
	a := m.transports[nodeA]
	b := m.transports[nodeB]
	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)

	m.StartAllTransports()
	initialNodes := m.Membership(nodeA, nodeB)
	t.Log(">>> membership")
	t.Log(initialNodes)

	t.Log(">>> connecting nodes")
	a.Connect(initialNodes)
	b.Connect(initialNodes)
	a.WaitFor(2)
	m.testEventuallyConnected(nodeA, nodeB)

	t.Log(">>> sending messages")
	nodeBEventsChan := b.EventsOut()
	err := a.Send(nodeB, testMsg.Pb())
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnectionsEmpty()

	m.CloseAllHosts()
	m.testNoHostConnections()
}

// TestNoConnectionsAfterStop tests that connections are not established after stop.
func TestNoConnectionsAfterStop(t *testing.T) {
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
	m.StopAllTransports()

	a.Connect(initialNodes)
	b.Connect(initialNodes)
	m.testNeverConnected(nodeA, nodeB)
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnectionsEmpty()
	m.CloseAllHosts()
	m.testNoHostConnections()
}

// TestSendReceiveWithWaitForAndBlock tests that the transport operates normally if WaitFor call and
// blocking messages are used.
func TestSendReceiveWithWaitForAndBlock(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	m := newMockLibp2pCommunication(t, DefaultParams(), []types.NodeID{nodeA, nodeB}, logger)

	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAllTransports()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)

	testMsg := checkpointpbmsgs.Checkpoint("", 0, 0, []byte{}, []byte{}) // Just a random message, could be anything.

	t.Log(">>> connecting nodes")

	initialNodes := m.Membership(nodeA, nodeB)

	t.Log("membership")
	t.Log(initialNodes)

	a.Connect(initialNodes)
	b.Connect(initialNodes)
	m.testEventuallyConnected(nodeA, nodeB)
	a.WaitFor(2)

	t.Log(">>> send a message")
	err := a.Send(nodeB, testMsg.Pb())
	require.NoError(t, err)

	nodeBEventsChan := b.incomingMessages

	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	t.Log(">>> blocking message")

	go func() {
		b.incomingMessages <- events.ListOf(
			transportpbevents.MessageReceived("1", "blocker", testMsg).Pb(),
		)
		b.incomingMessages <- events.ListOf(
			transportpbevents.MessageReceived("1", "blocker", testMsg).Pb(),
		)
	}()
	err = a.Send(nodeB, testMsg.Pb())
	time.Sleep(5 * time.Second)
	require.NoError(t, err)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA)
	m.testEventuallyNoStreams(nodeB)
	m.testConnectionsEmpty()

	m.CloseAllHosts()
	m.testNoHostConnections()

	m.testThatSenderIs(<-nodeBEventsChan, "blocker")
	m.testThatSenderIs(<-nodeBEventsChan, "blocker")
}

// TestOpeningConnectionAfterFail tests that when a node fails to connect first time it will try to connect again.
func TestOpeningConnectionAfterFail(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := types.NodeID("a")
	nodeB := types.NodeID("b")

	hostAddrA := libp2putil.NewFreeHostAddr()
	hostA := libp2putil.NewDummyHost(0, hostAddrA)
	addrA := types.NodeAddress(libp2putil.NewDummyMultiaddr(0, hostAddrA))
	a := NewTransport(DefaultParams(), nodeA, hostA, logger)

	hostAddrB := libp2putil.NewFreeHostAddr()
	hostB := libp2putil.NewDummyHostNoListen(1, hostAddrB)
	addrB := types.NodeAddress(libp2putil.NewDummyMultiaddr(1, hostAddrB))
	b := NewTransport(DefaultParams(), nodeB, hostB, logger)

	membershipA := make(map[types.NodeID]types.NodeAddress, 2)
	membershipA[nodeA] = addrA
	membershipA[nodeB] = addrB

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)

	m := &mockLibp2pCommunication{
		t:          t,
		params:     DefaultParams(),
		membership: make(map[types.NodeID]types.NodeAddress, 2),
		hosts:      make(map[types.NodeID]host.Host),
		transports: make(map[types.NodeID]*Transport),
		logger:     logger,
	}

	m.transports[nodeA] = a
	m.transports[nodeB] = b
	m.hosts[nodeA] = hostA
	m.hosts[nodeB] = hostB
	m.membership = membershipA

	t.Log(">>> connecting to a failed node")
	err := a.Start()
	require.NoError(t, err)
	err = b.Start()
	require.NoError(t, err)

	a.Connect(membershipA)
	m.testNeverConnected(nodeA, nodeB)
	b.Stop()
	m.CloseHosts(b)

	t.Log(">>> connecting to the restarted node")
	hostB = libp2putil.NewDummyHost(1, hostAddrB)
	b = NewTransport(DefaultParams(), nodeB, hostB, logger)
	m.transports[nodeB] = b
	err = b.Start()
	require.NoError(t, err)
	a.Connect(membershipA)
	m.testEventuallyHasConnection(nodeA, nodeB)

	t.Log(">>> cleaning")
	m.StopAllTransports()
	m.testEventuallyNoStreamsBetween(nodeA, nodeB)
	m.testEventuallyNoStreams(nodeA, nodeB)
	m.testConnectionsEmpty()

	m.CloseAllHosts()
	m.testNoHostConnections()
}

// TestMessagingWithNewNodes tests that the transport operates normally if new nodes are connected.
//
// This test connects N initial nodes with each other, starts sending messages between them.
// Then, in parallel we add m new nodes and connect them with all other nodes.
// Test checks that the implementation will not be in stuck and will not panic during those operations.
//
// nolint:gocognit
func TestMessagingWithNewNodes(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	N := 10 // number of nodes
	M := 5  // number of new nodes
	var nodes []types.NodeID
	for i := 0; i < N+M; i++ {
		nodes = append(nodes, types.NodeID(fmt.Sprint(i)))
	}

	m := newMockLibp2pCommunication(t, DefaultParams(), nodes, logger)
	m.StartAllTransports()
	initialNodes := m.Membership(nodes[:N]...)
	allNodes := m.Membership(nodes...)

	testMsg := checkpointpbmsgs.Checkpoint("", 0, 0, []byte{}, []byte{}) // Just a random message, could be anything.

	t.Logf(">>> connecting nodes")
	for i := 0; i < N; i++ {
		m.transports[nodes[i]].Connect(initialNodes)
	}

	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if i != j {
				m.testEventuallyConnected(nodes[i], nodes[j])
			}
		}
	}

	t.Log(">>> sending messages")

	var received, sent int64
	var senders, receivers sync.WaitGroup

	sender := func(src, dst types.NodeID) {
		defer senders.Done()

		for i := 0; i < 100; i++ {
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond) // nolint
			err := m.transports[src].Send(dst, testMsg.Pb())
			if err != nil {
				t.Logf("%v->%v failed to send: %v", src, dst, err)
			} else {
				atomic.AddInt64(&sent, 1)
			}
		}
	}

	receiver := func(nodeID types.NodeID, events chan *events.EventList, stop chan struct{}) {
		defer receivers.Done()

		for {
			select {
			case <-stop:
				return
			case e, ok := <-events:
				if !ok {
					return
				}
				_, valid := e.Iterator().Next().Type.(*eventpb.Event_Transport).Transport.Type.(*transportpb.Event_MessageReceived)
				require.Equal(m.t, true, valid)
				atomic.AddInt64(&received, 1)
			}
		}
	}

	for i := 0; i < N+M; i++ {
		for j := 0; j < N+M; j++ {
			if i != j {
				senders.Add(1)
				go sender(nodes[i], nodes[j])
			}
		}
	}

	for i := 0; i < N+M; i++ {
		ch := m.transports[nodes[i]]
		receivers.Add(1)
		go receiver(ch.ownID, ch.incomingMessages, ch.stop)
	}

	done := make(chan struct{})
	go func() {
		for i := 0; i < M+N; i++ {
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond) // nolint
			m.transports[nodes[i]].Connect(allNodes)
		}
		for i := 0; i < N+M; i++ {
			for j := 0; j < N+M; j++ {
				if i != j {
					m.testEventuallyConnected(nodes[i], nodes[j])
				}
			}
		}
		done <- struct{}{}
	}()

	senders.Wait()

	t.Log(">>> cleaning")
	<-done
	m.StopAllTransports()
	receivers.Wait()

	for _, n := range nodes {
		m.testEventuallyNoStreams(n)
	}
	m.testConnectionsEmpty()
	m.CloseAllHosts()
	m.testNoHostConnections()

	t.Logf(">>> sent: %d, received: %d", sent, received)
}
