package libp2p

import (
	"context"
	"sync"
	"time"

	es "github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	transportpbtypes "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Stats represents a statistics tracker for networking.
// It tracks the amount of data that is sent, received, and dropped (i.e., not sent).
// The data can be categorized using string labels, and it is the implementation's choice how to interpret them.
// All method implementations must be thread-save, as they can be called concurrently.
type Stats interface {
	Sent(nBytes int, label string)
	Received(nBytes int, label string)
	Dropped(nBytes int, label string)
}

type TransportMessage struct {
	Sender  string
	Payload []byte
}

type Transport struct {
	params Params
	ownID  t.NodeID
	host   host.Host
	logger logging.Logger
	stats  Stats

	connections     map[t.NodeID]connection
	nodeIDs         map[peer.ID]t.NodeID
	connectionsLock sync.RWMutex
	stop            chan struct{}
	incomingConnWg  sync.WaitGroup
	lastComplaint   time.Time

	incomingMessages chan *events.EventList
}

func NewTransport(params Params, ownID t.NodeID, h host.Host, logger logging.Logger, stats Stats) *Transport {
	return &Transport{
		params:           params,
		ownID:            ownID,
		host:             h,
		logger:           logger,
		connections:      make(map[t.NodeID]connection),
		nodeIDs:          make(map[peer.ID]t.NodeID),
		incomingMessages: make(chan *events.EventList),
		stop:             make(chan struct{}),
		stats:            stats,
	}
}

func (tr *Transport) ImplementsModule() {}

func (tr *Transport) ApplyEvents(_ context.Context, eventList *events.EventList) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		switch e := event.Type.(type) {
		case *eventpb.Event_Init:
			// no actions on init
		case *eventpb.Event_Transport:
			switch e := transportpbtypes.EventFromPb(e.Transport).Type.(type) {
			case *transportpbtypes.Event_SendMessage:
				for _, destID := range e.SendMessage.Destinations {
					if err := tr.Send(destID, e.SendMessage.Msg.Pb()); err != nil {

						// Complain if not complained recently.
						if time.Since(tr.lastComplaint) >= tr.params.MinComplainPeriod {
							tr.logger.Log(logging.LevelWarn, "Failed to send a message", "dest", destID, "err", err)
							tr.lastComplaint = time.Now()
						}

						// Update statistics about dropped message.
						if tr.stats != nil {
							var msgData []byte
							if msgData, err = encodeMessage(e.SendMessage.Msg.Pb(), tr.ownID); err != nil {
								return es.Errorf("failed to encode message (type %T)", e.SendMessage.Msg.Type)
							}
							tr.stats.Dropped(len(msgData), string(e.SendMessage.Msg.DestModule.Top()))
						}
					}
				}
			default:
				return es.Errorf("unexpected transport event: %T", e)
			}
		default:
			return es.Errorf("unexpected event: %T", event.Type)
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

	// Stop receiving new connections.
	// After the lock is released, no new incoming connections can be created anymore.
	tr.connectionsLock.Lock()
	tr.host.RemoveStreamHandler(tr.params.ProtocolID)
	close(tr.stop)
	tr.connectionsLock.Unlock()

	// Passing an empty membership means that no connections will be kept.
	tr.CloseOldConnections(&trantorpbtypes.Membership{map[t.NodeID]*trantorpbtypes.NodeIdentity{}}) // nolint:govet

	// Wait for incoming connection handlers to terminate.
	tr.incomingConnWg.Wait()
}

func (tr *Transport) Connect(membership *trantorpbtypes.Membership) {
	tr.connectionsLock.Lock()
	defer tr.connectionsLock.Unlock()

	select {
	case <-tr.stop:
		tr.logger.Log(logging.LevelWarn, "called connect after stop")
		return
	default:
	}

	for nodeID, identity := range membership.Nodes {
		// For each given node

		if _, ok := tr.connections[nodeID]; !ok {
			// If connection doesn't yet exist

			// Parse node address.
			addr, err := multiaddr.NewMultiaddr(identity.Addr)
			if err != nil {
				tr.logger.Log(logging.LevelError, "Failed parsing address. Ignoring.", "addr", identity.Addr)
				continue
			}

			// Create a new connection.
			// This is a non-blocking call - all I/O will be done by the connection in the background.
			// A connection to self needs a special treatment, as libp2p does not allow to dial oneself.
			var conn connection
			if nodeID == tr.ownID {
				conn, err = newSelfConnection(tr.params, tr.ownID, addr, tr.incomingMessages)
			} else {
				conn, err = newRemoteConnection(
					tr.params,
					tr.ownID,
					addr,
					tr.host,
					logging.Decorate(tr.logger, "SND: ", "dest", nodeID),
					tr.stats,
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

func (tr *Transport) WaitFor(n int) error {

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
				return err
			}
			numConnected++
		case <-tr.stop:
			return es.Errorf("transport stopped while waiting for connections")
		}
	}

	return nil
}

func (tr *Transport) CloseOldConnections(newMembership *trantorpbtypes.Membership) {

	// Select connections to nodes that are NOT part of the new membership.
	// We first select the connections (while holding a lock) and only then close them (after releasing the lock),
	// As closing a connection involves waiting for its processing goroutine to stop.
	toClose := make(map[t.NodeID]connection)
	tr.connectionsLock.Lock()
	for nodeID, conn := range tr.connections {
		if _, ok := newMembership.Nodes[nodeID]; !ok {
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
		return nil, es.Errorf("no connection to node: %v", nodeID)
	}
	return conn, nil
}

func (tr *Transport) handleIncomingConnection(s network.Stream) {
	// In case there is a panic in the main processing loop, log an error message.
	// (Otherwise, since this function is run as a goroutine, panicking would be completely silent.)
	defer func() {
		if r := recover(); r != nil {
			err := es.New(r)
			tr.logger.Log(logging.LevelError, "Incoming connection handler panicked.",
				"cause", r, "stack", err.ErrorStack())
		}
	}()

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

	tr.connectionsLock.RLock()

	// Check if connection has been received from a known node (as per ID declared by the remote node).
	// (The actual check is performed after we release the lock.)
	nodeID, ok := tr.nodeIDs[peerID]

	select {
	case <-tr.stop:
		// If we are shutting down right now, ignore this connection.
		return
	default:
		// If new connections are still accepted, make sure we will wait for this handler to terminate when stopping.
		tr.incomingConnWg.Add(1)
		defer tr.incomingConnWg.Done()
	}

	tr.connectionsLock.RUnlock()

	if !ok {
		tr.logger.Log(logging.LevelWarn, "Received message from unknown peer. Stopping incoming connection",
			"remotePeer", peerID)
		return
	}

	// Create a goroutine watching for the stop signal (closing tr.stop) and closing the connection on shutdown.
	// It is necessary for the case where the handler is blocked on tr.readAndProcessMessages and needs to shut down.
	// TODO: Resetting and closing the connection is done both in this goroutine and when the handler returns.
	//   This might be redundant and is, in general, a mess. Clean it up.
	connStopChan := make(chan struct{})
	defer close(connStopChan)
	go func() {
		select {
		case <-tr.stop:
			if err := s.Reset(); err != nil {
				tr.logger.Log(logging.LevelWarn, "Could not reset incoming stream on shutdown.",
					"remotePeer", peerID, "err", err)
			}
			if err := s.Close(); err != nil {
				tr.logger.Log(logging.LevelWarn, "Could not close incoming stream on shutdown.",
					"remotePeer", peerID, "err", err)
			}
		case <-connStopChan:
			// Do nothing here.
			// This is only used to return from this goroutine if the connection handler stops by itself.
		}
	}()

	// Start reading and processig incoming messages.
	// This call blocks until an error occurs (e.g. the connection breaks) or the Transport is explicitly shut down.
	tr.readAndProcessMessages(s, nodeID, peerID)
}

// readAndProcessMessages Reads incoming data from the stream s, parses it to Mir messages,
// and writes them to the incomingMessages channel
func (tr *Transport) readAndProcessMessages(s network.Stream, nodeID t.NodeID, peerID peer.ID) {
	for {
		// Read message from the network.
		msg, sender, err := readAndDecode(s, tr.stats)
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
		case tr.incomingMessages <- events.ListOf(transportpbevents.MessageReceived(
			t.ModuleID(msg.DestModule),
			sender,
			messagepbtypes.MessageFromPb(msg),
		).Pb()):
			// Nothing to do in this case message has written to the receiving channel.
		case <-tr.stop:
			tr.logger.Log(logging.LevelError, "Shutdown. Stopping incoming connection.",
				"nodeID", sender, "remotePeer", peerID)
			return
		}
	}
}

func readAndDecode(s network.Stream, stats Stats) (*messagepb.Message, t.NodeID, error) {
	var tm TransportMessage
	err := tm.UnmarshalCBOR(s)
	if err != nil {
		return nil, "", err
	}

	var msg messagepb.Message
	if err := proto.Unmarshal(tm.Payload, &msg); err != nil {
		return nil, "", err
	}

	// Record statistics about received data.
	// TODO: There is an asymmetry between the recorded sent data size and the received one:
	//   When sending we record the size of the CBOR serialized data (thus including the CBOR serialization overhead),
	//   but we only record the size of the deserialized elements of the CBOR-serializable message type.
	//   Thus, in the end result the amount of data received will be less (by the CBOR serialization overhead)
	//   than the amount of data sent. The difference is expected to be marginal though.
	if stats != nil {
		stats.Received(len(tm.Payload)+len(tm.Sender), string(t.ModuleID(msg.DestModule).Top()))
	}

	return &msg, t.NodeID(tm.Sender), nil
}
