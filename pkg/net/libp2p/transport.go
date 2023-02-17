package libp2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type TransportMessage struct {
	Sender  string
	Payload []byte
}

type Transport struct {
	params Params
	ownID  t.NodeID
	host   host.Host
	logger logging.Logger

	connections     map[t.NodeID]connection
	nodeIDs         map[peer.ID]t.NodeID
	connectionsLock sync.RWMutex
	stop            chan struct{}

	incomingMessages chan *events.EventList
}

func NewTransport(params Params, ownID t.NodeID, h host.Host, logger logging.Logger) *Transport {
	return &Transport{
		params:           params,
		ownID:            ownID,
		host:             h,
		logger:           logger,
		connections:      make(map[t.NodeID]connection),
		nodeIDs:          make(map[peer.ID]t.NodeID),
		incomingMessages: make(chan *events.EventList),
		stop:             make(chan struct{}),
	}
}

func (tr *Transport) ImplementsModule() {}

func (tr *Transport) ApplyEvents(_ context.Context, eventList *events.EventList) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		switch e := event.Type.(type) {
		case *eventpb.Event_Init:
			// no actions on init
		case *eventpb.Event_SendMessage:
			for _, destID := range e.SendMessage.Destinations {
				if err := tr.Send(t.NodeID(destID), e.SendMessage.Msg); err != nil {
					tr.logger.Log(logging.LevelWarn, "Failed to send a message", "dest", destID, "err", err)
				}
			}
		default:
			return fmt.Errorf("unexpected event: %T", event.Type)
		}
	}

	return nil
}

func (tr *Transport) EventsOut() <-chan *events.EventList {
	return tr.incomingMessages
}

func (tr *Transport) Start() error {
	tr.host.SetStreamHandler(tr.params.ProtocolID, tr.handleIncomingConnection)
	return nil
}

func (tr *Transport) Stop() {
	tr.host.RemoveStreamHandler(tr.params.ProtocolID)
	close(tr.stop)
	tr.CloseOldConnections(map[t.NodeID]t.NodeAddress{}) // Passing an empty map means that no connections will be kept.
	// TODO: Force the termination of incoming connections too and wait here until that happens.
}

func (tr *Transport) Connect(nodes map[t.NodeID]t.NodeAddress) {
	tr.connectionsLock.Lock()
	defer tr.connectionsLock.Unlock()

	select {
	case <-tr.stop:
		tr.logger.Log(logging.LevelWarn, "called connect after stop")
		return
	default:
	}

	for nodeID, addr := range nodes {
		// For each given node

		if _, ok := tr.connections[nodeID]; !ok {
			// If connection doesn't yet exist

			// Create a new connection.
			// This is a non-blocking call - all I/O will be done by the connection in the background.
			// A connection to self needs a special treatment, as libp2p does not allow to dial oneself.
			var conn connection
			var err error
			if nodeID == tr.ownID {
				conn, err = newSelfConnection(tr.params, tr.ownID, addr, tr.incomingMessages)
			} else {
				conn, err = newRemoteConnection(
					tr.params,
					tr.ownID,
					addr,
					tr.host,
					logging.Decorate(tr.logger, "SND: ", "dest", nodeID),
				)
			}

			// Complain and ignore node if connection cannot be created.
			// Note that this is different from not (yet) establishing an underlying network connection,
			// which does not result in an error and will be retried in the background.
			if err != nil {
				tr.logger.Log(logging.LevelError, "Cannot create new connection, ignoring destination.",
					"err", err, "nodeID", nodeID, "addr", addr)
			} else {
				tr.connections[nodeID] = conn
				tr.nodeIDs[conn.PeerID()] = nodeID
			}
		}
	}
}

func (tr *Transport) Send(dest t.NodeID, msg *messagepb.Message) error {
	conn, err := tr.getConnection(dest)
	if err != nil {
		return err
	}
	return conn.Send(msg)
}

func (tr *Transport) WaitFor(n int) {

	tr.connectionsLock.RLock()

	// We store the cancel functions returned by all calls to conn.Wait().
	// When done waiting (for whatever reason), call all those functions
	// to unblock the goroutines that are still waiting.
	cancelFuncs := make([]func(), 0, len(tr.connections))
	defer func() {
		for _, cancelFunc := range cancelFuncs {
			cancelFunc()
		}
	}()

	// Invoke Wait() on each connection and have the value it outputs written to the aggregator channel.
	counterChan := make(chan error) // All the outputs produced by conn.Wait() calls are aggregated here.
	for _, conn := range tr.connections {
		errChan, cancelFunc := conn.Wait()
		cancelFuncs = append(cancelFuncs, cancelFunc)
		go func() {
			counterChan <- <-errChan // Read from one channel and directly write the value to another (the aggregator).
		}()
	}

	tr.connectionsLock.RUnlock()

	// Read values from the counter channel (aggregator) until enough successful
	// connections (nil error) have been reached or an error occurs.
	numConnected := 0
	for numConnected < n {
		select {
		case err := <-counterChan:
			if err != nil {
				return // TODO: Return an error here (adaptation of the Transport interface needed)
			}
			numConnected++
		case <-tr.stop:
			return // TODO: Return the error here too.
		}
	}
}

func (tr *Transport) CloseOldConnections(newMembership map[t.NodeID]t.NodeAddress) {

	// Select connections to nodes that are NOT part of the new membership.
	// We first select the connections (while holding a lock) and only then close them (after releasing the lock),
	// As closing a connection involves waiting for its processing goroutine to stop.
	toClose := make(map[t.NodeID]connection)
	tr.connectionsLock.Lock()
	for nodeID, conn := range tr.connections {
		if _, ok := newMembership[nodeID]; !ok {
			toClose[nodeID] = conn
			delete(tr.nodeIDs, conn.PeerID())
			delete(tr.connections, nodeID)
		}
	}
	tr.connectionsLock.Unlock()

	// Close the selected connections (in parallel).
	wg := sync.WaitGroup{}
	wg.Add(len(toClose))
	for nodeID, conn := range toClose {
		go func(nodeID t.NodeID, conn connection) {
			conn.Close()
			tr.logger.Log(logging.LevelInfo, "Closed connection to node.", "nodeID", nodeID)
			wg.Done()
		}(nodeID, conn)
	}
	wg.Wait()
}

func (tr *Transport) getConnection(nodeID t.NodeID) (connection, error) {
	tr.connectionsLock.RLock()
	defer tr.connectionsLock.RUnlock()

	conn, ok := tr.connections[nodeID]
	if !ok {
		return nil, fmt.Errorf("no connection to node: %v", nodeID)
	}
	return conn, nil
}

func (tr *Transport) handleIncomingConnection(s network.Stream) {
	peerID := s.Conn().RemotePeer()

	tr.logger.Log(logging.LevelDebug, "Incoming connection", "remotePeer", peerID)

	// Close the stream when done.
	defer func() {
		if err := s.Reset(); err != nil {
			tr.logger.Log(logging.LevelWarn, "Could not reset incoming stream", "remotePeer", peerID, "err", err)
		}
		if err := s.Close(); err != nil {
			tr.logger.Log(logging.LevelWarn, "Could not close incoming stream", "remotePeer", peerID, "err", err)
		}
	}()

	// Check if connection has been received from a known node (as per ID declared by the remote node).
	tr.connectionsLock.RLock()
	nodeID, ok := tr.nodeIDs[peerID]
	// TODO: Also keep a synchronized map of incoming streams.
	//       On shutdown, close them all and wait until the corresponding handlers return.
	tr.connectionsLock.RUnlock()
	if !ok {
		tr.logger.Log(logging.LevelWarn, "Received message from unknown peer. Stopping incoming connection",
			"remotePeer", peerID)
		return
	}

	// Start reading and processig incoming messages.
	// This call blocks until an error occurs (e.g. the connection breaks) or the Transport is explicitly shut down.
	tr.readAndProcessMessages(s, nodeID, peerID)
}

// readAndProcessMessages Reads incoming data from the stream s, parses it to Mir messages,
// and writes them to the incomingMessages channel
func (tr *Transport) readAndProcessMessages(s network.Stream, nodeID t.NodeID, peerID peer.ID) {
	for {
		// Read message from the network.
		msg, sender, err := readAndDecode(s)
		if err != nil {
			tr.logger.Log(logging.LevelDebug, "Failed reading message. Stopping incoming connection",
				"remotePeer", peerID, "err", err)
			return
		}

		// Sanity check. TODO: Remove the `sender` completely and infer it from the PeerID.
		if sender != nodeID {
			tr.logger.Log(logging.LevelWarn, "Remote node identity mismatch. Stopping incoming connection.",
				"expectedNodeID", nodeID, "declaredNodeID", sender, "remotePeerID", peerID)
			return
		}

		select {
		// TODO: Think of a smarter way of batching the incoming messages on the channel,
		//       instead of sending each individual message as a list of length one.
		case tr.incomingMessages <- events.ListOf(
			events.MessageReceived(t.ModuleID(msg.DestModule), sender, msg),
		):
		case <-tr.stop:
			tr.logger.Log(logging.LevelError, "Shutdown. Stopping incoming connection.", "nodeID", sender, "remotePeer", peerID)
			return
		}
	}
}

func readAndDecode(s network.Stream) (*messagepb.Message, t.NodeID, error) {
	var tm TransportMessage
	err := tm.UnmarshalCBOR(s)
	if err != nil {
		return nil, "", err
	}

	var msg messagepb.Message
	if err := proto.Unmarshal(tm.Payload, &msg); err != nil {
		return nil, "", err
	}
	return &msg, t.NodeID(tm.Sender), nil
}
