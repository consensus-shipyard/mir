package grpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	"github.com/filecoin-project/mir/pkg/pb/transportpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

type mockGrpcCommunication struct {
	t          *testing.T
	params     Params
	membership *trantorpbtypes.Membership
	logger     logging.Logger
	transports map[stdtypes.NodeID]*Transport
}

func newMockGrpcCommunication(
	t *testing.T,
	params Params,
	nodeIDs []stdtypes.NodeID,
	logger logging.Logger,
) *mockGrpcCommunication {
	lt := &mockGrpcCommunication{
		t:      t,
		params: params,
		membership: &trantorpbtypes.Membership{ // nolint:govet
			make(map[stdtypes.NodeID]*trantorpbtypes.NodeIdentity, len(nodeIDs)),
		},
		transports: make(map[stdtypes.NodeID]*Transport),
		logger:     logger,
	}

	for i, id := range nodeIDs {
		lt.membership.Nodes[id] = &trantorpbtypes.NodeIdentity{ // nolint:govet
			id,
			mockAddr(i),
			nil,
			"1",
		}
		var err error
		lt.transports[id], err = NewTransport(params, id, lt.membership.Nodes[id].Addr, logging.Decorate(logger, fmt.Sprintf("%v: ", id)), nil)
		require.NoError(t, err)
	}

	return lt
}

func mockAddr(i int) string {
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 10000+i)
}

func (m *mockGrpcCommunication) getTransport(sourceID stdtypes.NodeID) *Transport {
	tr, ok := m.transports[sourceID]
	if !ok {
		m.t.Fatalf("failed to get transport for node: %v", sourceID)
	}
	return tr
}

func (m *mockGrpcCommunication) Membership() *trantorpbtypes.Membership {
	return m.membership
}

func (m *mockGrpcCommunication) StopTransports(transports ...*Transport) {
	for _, v := range transports {
		v.Stop()
	}
}

func (m *mockGrpcCommunication) StopAllTransports() {
	for _, v := range m.transports {
		v.Stop()
	}
}

func (m *mockGrpcCommunication) StartTransports(transports ...*Transport) {
	for _, v := range transports {
		err := v.Start()
		require.NoError(m.t, err, "starting node", v.ownID)
	}
}

func (m *mockGrpcCommunication) StartAllTransports() {
	for _, v := range m.transports {
		err := v.Start()
		require.NoError(m.t, err, "starting node", v.ownID)
	}
}

func (m *mockGrpcCommunication) MembershipOf(ids ...stdtypes.NodeID) *trantorpbtypes.Membership {
	membership := &trantorpbtypes.Membership{make(map[stdtypes.NodeID]*trantorpbtypes.NodeIdentity)} // nolint:govet
	for _, id := range ids {
		identity, ok := m.membership.Nodes[id]
		if !ok {
			m.t.Fatalf("failed to get addr for a new node %v", id)
		}
		membership.Nodes[id] = identity
	}
	return membership
}

func (m *mockGrpcCommunication) FourTransports(nodeID ...stdtypes.NodeID) (*Transport, *Transport, *Transport, *Transport) {
	if len(nodeID) != 4 {
		m.t.Fatalf("want 4 node IDs, have %d", len(nodeID))
	}

	ts := [4]*Transport{}

	for i, v := range nodeID {
		ts[i] = m.transports[v]
	}

	return ts[0], ts[1], ts[2], ts[3]
}

func (m *mockGrpcCommunication) testThatSenderIs(events *stdtypes.EventList, nodeID stdtypes.NodeID) {
	switch event := events.Iterator().Next().(type) {
	case *eventpb.Event:
		tEvent, valid := event.Type.(*eventpb.Event_Transport)
		require.True(m.t, valid)
		msg, valid := tEvent.Transport.Type.(*transportpb.Event_MessageReceived)
		require.True(m.t, valid)
		require.Equal(m.t, string(nodeID), msg.MessageReceived.From)
	case *stdevents.MessageReceived:
		require.Equal(m.t, nodeID, event.Sender)
	default:
		panic(fmt.Sprintf("unknown event type: %T", event))
	}

}

func (m *mockGrpcCommunication) testConnected(nodeID1, nodeID2 stdtypes.NodeID, timeout time.Duration) {
	n1 := m.getTransport(nodeID1)
	n2 := m.getTransport(nodeID2)

	require.Eventually(m.t,
		func() bool {
			n1.connectionsLock.RLock()
			defer n1.connectionsLock.RUnlock()
			n2.connectionsLock.RLock()
			defer n2.connectionsLock.RUnlock()

			if n1.connections[nodeID2] == nil || n2.connections[nodeID1] == nil {
				return false
			}

			n1.connections[nodeID2].(*remoteConnection).connectedCond.L.Lock()
			defer n1.connections[nodeID2].(*remoteConnection).connectedCond.L.Unlock()
			n2.connections[nodeID1].(*remoteConnection).connectedCond.L.Lock()
			defer n2.connections[nodeID1].(*remoteConnection).connectedCond.L.Unlock()

			return n1.connections[nodeID2].(*remoteConnection).msgSink != nil &&
				n2.connections[nodeID1].(*remoteConnection).msgSink != nil
		},
		timeout, 200*time.Millisecond)
}

// TestNewTransport tests that the transport can be created.
func TestNewTransport(t *testing.T) {
	_, err := NewTransport(DefaultParams(), "a", mockAddr(0), logging.ConsoleDebugLogger, nil)
	require.NoError(t, err)
}

// TestTransportStartStop tests that the transport can be started and stopped.
func TestTransportStartStop(t *testing.T) {
	tr, err := NewTransport(DefaultParams(), "a", mockAddr(0), logging.ConsoleDebugLogger, nil)
	require.NoError(t, err)

	err = tr.Start()
	require.NoError(t, err)
	tr.Stop()
}

// TestConnectTwoNodes that two nodes can be connected.
func TestConnectTwoNodes(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	nodeA := stdtypes.NodeID("a")
	nodeB := stdtypes.NodeID("b")
	m := newMockGrpcCommunication(t, DefaultParams(), []stdtypes.NodeID{nodeA, nodeB}, logger)
	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAllTransports()

	t.Log(">>> membership")
	initialNodes := m.MembershipOf(nodeA, nodeB)
	t.Log(initialNodes)

	t.Log(">>> connecting nodes")
	a.Connect(initialNodes)
	b.Connect(initialNodes)

	m.testConnected(nodeA, nodeB, 10*time.Second)

	m.StopAllTransports()
}

// TestSendReceive tests that messages can be sent and received.
func TestSendReceive(t *testing.T) {
	logger := logging.ConsoleDebugLogger

	testMsg := stdtypes.RawMessage([]byte{0, 1, 2, 3})
	testPbMsg := &messagepbtypes.Message{}

	nodeA := stdtypes.NodeID("a")
	nodeB := stdtypes.NodeID("b")
	m := newMockGrpcCommunication(t, DefaultParams(), []stdtypes.NodeID{nodeA, nodeB}, logger)
	a := m.transports[nodeA]
	b := m.transports[nodeB]
	m.StartAllTransports()

	require.Equal(t, nodeA, a.ownID)
	require.Equal(t, nodeB, b.ownID)

	membership := m.Membership()

	t.Log(">>> connecting nodes")
	a.Connect(membership)
	b.Connect(membership)

	m.testConnected(nodeA, nodeB, 30*time.Second)

	t.Log(">>> sending messages")

	nodeAEventsChan := a.EventsOut()
	nodeBEventsChan := b.EventsOut()

	err := a.SendPbMessage(nodeB, testPbMsg.Pb())
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)
	err = a.SendRawMessage(nodeB, "dummyModule", testMsg)
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeBEventsChan, nodeA)

	err = b.SendPbMessage(nodeA, testPbMsg.Pb())
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeAEventsChan, nodeB)
	err = b.SendRawMessage(nodeA, "dummyModule", testMsg)
	require.NoError(t, err)
	m.testThatSenderIs(<-nodeAEventsChan, nodeB)

	t.Log("Stopping all transports.")

	m.StopAllTransports()
}
