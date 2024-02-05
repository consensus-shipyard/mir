package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	es "github.com/go-errors/errors"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/logging"
	mirnet "github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	transportpbtypes "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

type TransportMessage struct {
	Sender  []byte
	Payload []byte
}

type Transport struct {
	UnimplementedGrpcTransportServer

	// The address of the node.
	ownAddr stdtypes.NodeAddress

	// The gRPC server used by this networking module.
	grpcServer *grpc.Server

	// Error returned from the grpcServer.Serve() call (see Start() method).
	grpcServerError error

	params Params
	ownID  stdtypes.NodeID
	logger logging.Logger
	stats  mirnet.Stats

	connections     map[stdtypes.NodeID]connection
	nodeIDs         map[string]stdtypes.NodeID
	connectionsLock sync.RWMutex
	stop            chan struct{}
	lastComplaint   time.Time

	incomingMessages chan *stdtypes.EventList
}

func NewTransport(
	params Params,
	ownID stdtypes.NodeID,
	ownAddr string,
	logger logging.Logger,
	stats mirnet.Stats,
) (*Transport, error) {
	addr, err := stdtypes.NodeAddressFromString(ownAddr)
	if err != nil {
		return nil, es.Errorf("failed parsing own address: %w", err)
	}

	return &Transport{
		ownAddr:          addr,
		ownID:            ownID,
		params:           params,
		logger:           logger,
		connections:      make(map[stdtypes.NodeID]connection),
		nodeIDs:          make(map[string]stdtypes.NodeID),
		incomingMessages: make(chan *stdtypes.EventList),
		stop:             make(chan struct{}),
		stats:            stats,
	}, nil
}

func (tr *Transport) ImplementsModule() {}

func (tr *Transport) ApplyEvents(
	ctx context.Context,
	eventList *stdtypes.EventList,
) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		switch evt := event.(type) {
		case *stdevents.Init:
			// no actions on init
		case *stdevents.SendMessage:
			for _, destID := range evt.DestNodes {
				if destID == tr.ownID {
					// Send message to myself bypassing the network.
					// The sending must be done in its own goroutine in case writing to tr.incomingMessages blocks.
					// (Processing of input events must be non-blocking.)
					receiveEvent := stdevents.NewMessageReceived(evt.RemoteDestModule, tr.ownID, evt.Payload)
					go func() {
						select {
						case tr.incomingMessages <- stdtypes.ListOf(receiveEvent):
						case <-ctx.Done():
						}
					}()
				} else {
					// Send message to another node.
					if err := tr.SendRawMessage(destID, evt.RemoteDestModule, evt.Payload); err != nil {

						// Complain if not complained recently.
						if time.Since(tr.lastComplaint) >= tr.params.MinComplainPeriod {
							tr.logger.Log(logging.LevelWarn, "Failed to send a message", "dest", destID, "err", err)
							tr.lastComplaint = time.Now()
						}

						// Update statistics about dropped message.
						if tr.stats != nil {
							msgData, err := evt.Payload.ToBytes()
							if err != nil {
								return es.Errorf("failed serializing message data: %w", err)
							}
							tr.stats.Dropped(proto.Size(&GrpcMessage{
								Sender: tr.ownID.Bytes(),
								Type: &GrpcMessage_RawMsg{RawMsg: &RawMessage{
									DestModule: evt.DestModule.String(), // TODO: Get rid of the String conversion here.
									Data:       msgData,
								}},
							}), string(evt.DestModule.Top()))
						}
					}
				}
			}
		case *eventpb.Event:
			return tr.ApplyPbEvent(ctx, evt)
		default:
			return es.Errorf("GRPC transport only supports proto events and OutgoingMessage, received %T", event)
		}
	}

	return nil
}

func (tr *Transport) ApplyPbEvent(ctx context.Context, evt *eventpb.Event) error {
	switch e := evt.Type.(type) {
	case *eventpb.Event_Transport:
		switch e := transportpbtypes.EventFromPb(e.Transport).Type.(type) {
		case *transportpbtypes.Event_SendMessage:
			for _, destID := range e.SendMessage.Destinations {
				if destID == tr.ownID {
					// Send message to myself bypassing the network.
					// The sending must be done in its own goroutine in case writing to tr.incomingMessages blocks.
					// (Processing of input events must be non-blocking.)
					receivedEvent := transportpbevents.MessageReceived(
						e.SendMessage.Msg.DestModule,
						tr.ownID,
						e.SendMessage.Msg,
					)
					go func() {
						select {
						case tr.incomingMessages <- stdtypes.ListOf(receivedEvent.Pb()):
						case <-ctx.Done():
						}
					}()
				} else {
					// Send message to another node.
					if err := tr.SendPbMessage(destID, e.SendMessage.Msg.Pb()); err != nil {

						// Complain if not complained recently.
						if time.Since(tr.lastComplaint) >= tr.params.MinComplainPeriod {
							tr.logger.Log(logging.LevelWarn, "Failed to send a message", "dest", destID, "err", err)
							tr.lastComplaint = time.Now()
						}

						// Update statistics about dropped message.
						if tr.stats != nil {
							tr.stats.Dropped(proto.Size(&GrpcMessage{
								Sender: tr.ownID.Bytes(),
								Type:   &GrpcMessage_PbMsg{PbMsg: e.SendMessage.Msg.Pb()},
							}), string(e.SendMessage.Msg.DestModule.Top()))
						}
					}
				}
			}
		default:
			return es.Errorf("unexpected type of transport event: %T", e)
		}
	default:
		return es.Errorf("unexpected type of Net event: %T", evt.Type)
	}

	return nil
}

func (tr *Transport) EventsOut() <-chan *stdtypes.EventList {
	return tr.incomingMessages
}

// Start starts the networking module by initializing and starting the internal gRPC server,
// listening on the port determined by the membership and own ID.
// Before ths method is called, no other GrpcTransports can connect to this one.
func (tr *Transport) Start() error {
	// Obtain net.Dial compatible address.
	_, dialAddr, err := manet.DialArgs(tr.ownAddr)
	if err != nil {
		return es.Errorf("failed to obtain Dial address: %w", err)
	}
	// Obtain own port number from membership.
	_, ownPort, err := net.SplitHostPort(dialAddr)
	if err != nil {
		return err
	}

	// Create a gRPC server and assign it the logic of this module.
	tr.grpcServer = grpc.NewServer()
	RegisterGrpcTransportServer(tr.grpcServer, tr)

	// Start listening on the network
	conn, err := net.Listen("tcp", ":"+ownPort)
	if err != nil {
		return es.Errorf("failed to listen for connections on port %s: %w", ownPort, err)
	}

	tr.logger.Log(logging.LevelInfo, fmt.Sprintf("Listening for connections on port %s", ownPort))

	// Start the gRPC server in a separate goroutine.
	// When the server stops, it will write its exit error into tr.grpcServerError.
	go func() {
		tr.grpcServerError = tr.grpcServer.Serve(conn)
	}()

	// If we got all the way here, no error occurred.
	return nil
}

// Stop closes all open connections to other nodes and stops the own gRPC server
// (preventing further incoming connections).
// After Stop() returns, the error returned by the gRPC server's Serve() call
// can be obtained through the ServerError() method.
func (tr *Transport) Stop() {

	// Stop receiving new connections.
	// After the lock is released, no new incoming connections can be created anymore.
	tr.connectionsLock.Lock()
	close(tr.stop)
	tr.connectionsLock.Unlock()

	// Passing an empty membership means that no connections will be kept.
	tr.CloseOldConnections(&trantorpbtypes.Membership{map[stdtypes.NodeID]*trantorpbtypes.NodeIdentity{}}) // nolint:govet

	tr.logger.Log(logging.LevelDebug, "Outgoing connections closed.")

	// Stop own gRPC server.
	tr.logger.Log(logging.LevelDebug, "Stopping gRPC server")
	tr.grpcServer.Stop()
	tr.logger.Log(logging.LevelDebug, "gRPC server stopped")
}

// ServerError returns the error returned by the gRPC server's Serve() call.
// ServerError() must not be called before the GrpcTransport is stopped and its Stop() method has returned.
func (tr *Transport) ServerError() error {
	return tr.grpcServerError
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

			_, dialAddr, err := manet.DialArgs(addr)
			if err != nil {
				tr.logger.Log(logging.LevelWarn, "Failed formatting address %s: %w", addr.String(), err)
			}

			// Create a new connection.
			// This is a non-blocking call - all I/O will be done by the connection in the background.
			// A connection to self circumvents the network completely.
			var conn connection
			if nodeID == tr.ownID {
				conn, err = newSelfConnection(tr.params, dialAddr, tr.incomingMessages)
			} else {
				conn, err = newRemoteConnection(
					tr.params,
					dialAddr,
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
				tr.nodeIDs[conn.Address()] = nodeID
			}
		}
	}
}

// SendPbMessage sends a protobuf type message msg to the node with ID dest.
// Concurrent calls to Send are not (yet? TODO) supported.
func (tr *Transport) SendPbMessage(destNode stdtypes.NodeID, msg *messagepb.Message) error {
	conn, err := tr.getConnection(destNode)
	if err != nil {
		return err
	}
	return conn.Send(&GrpcMessage{
		Sender: tr.ownID.Bytes(),
		Type:   &GrpcMessage_PbMsg{PbMsg: msg},
	})
}

func (tr *Transport) SendRawMessage(destNode stdtypes.NodeID, destModule stdtypes.ModuleID, message stdtypes.Message) error {
	conn, err := tr.getConnection(destNode)
	if err != nil {
		return err
	}

	msgData, err := message.ToBytes()
	if err != nil {
		return es.Errorf("failed serializing message data: %w", err)
	}

	return conn.Send(&GrpcMessage{
		Sender: tr.ownID.Bytes(),
		Type: &GrpcMessage_RawMsg{RawMsg: &RawMessage{
			DestModule: destModule.String(), // TODO: Get rid of the String conversion here.
			Data:       msgData,
		}},
	})
}

// Send sends a protobuf type message msg to the node with ID dest.
// Concurrent calls to Send are not (yet? TODO) supported.
// This is just a wrapper around SendPbMessage. TODO: generalize interface to take a stdtypes.Message instead.
func (tr *Transport) Send(dest stdtypes.NodeID, msg *messagepb.Message) error {
	return tr.SendPbMessage(dest, msg)
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
	toClose := make(map[stdtypes.NodeID]connection)
	tr.connectionsLock.Lock()
	for nodeID, conn := range tr.connections {
		if _, ok := newMembership.Nodes[nodeID]; !ok {
			toClose[nodeID] = conn
			delete(tr.nodeIDs, conn.Address())
			delete(tr.connections, nodeID)
		}
	}
	tr.connectionsLock.Unlock()

	// Close the selected connections (in parallel).
	wg := sync.WaitGroup{}
	wg.Add(len(toClose))
	for nodeID, conn := range toClose {
		go func(nodeID stdtypes.NodeID, conn connection) {
			conn.Close()
			tr.logger.Log(logging.LevelInfo, "Closed connection to node.", "nodeID", nodeID)
			wg.Done()
		}(nodeID, conn)
	}
	wg.Wait()
}

// Listen implements the gRPC Listen service (multi-request-single-response).
// It receives messages from the gRPC client running on the other node
// and writes them to a channel that the user can access through ReceiveChan().
// This function is called by the gRPC system on every new connection
// from another node's Net module's gRPC client.
func (tr *Transport) Listen(srv GrpcTransport_ListenServer) error {

	// In case there is a panic in the main processing loop, log an error message.
	// (Otherwise, since this function is run as a goroutine, panicking would be completely silent.)
	defer func() {
		if r := recover(); r != nil {
			err := es.New(r)
			tr.logger.Log(logging.LevelError, "Incoming connection handler panicked.",
				"cause", r, "stack", err.ErrorStack())
		}
	}()

	// Print address of incoming connection.
	p, ok := peer.FromContext(srv.Context())
	peerAddr := p.Addr.String()
	if ok {
		tr.logger.Log(logging.LevelDebug, fmt.Sprintf("Incoming connection from %s", peerAddr))
	} else {
		return es.Errorf("failed to get grpc peer info from context")
	}

	// Makes sure to close an incoming connection only once.
	var closeOnce sync.Once
	closeFunc := func() {
		tr.logger.Log(logging.LevelWarn, "Done handling incoming connection.", "addr", peerAddr)
		if err := srv.SendAndClose(&ByeBye{}); err != nil {
			tr.logger.Log(logging.LevelWarn, "Could not close incoming connection", "remotePeer", peerAddr, "err", err)
		}
	}

	// Close the stream when done.
	defer func() {
		closeOnce.Do(closeFunc)
	}()

	tr.connectionsLock.RLock()

	// Check if connection has been received from a known node.
	// (The actual check is performed after we release the lock.)
	_, ok = tr.nodeIDs[peerAddr]

	select {
	case <-tr.stop:
		// If we are shutting down right now, ignore this connection.
		tr.connectionsLock.RUnlock()
		return nil
	default:
		// Do nothing.
	}

	tr.connectionsLock.RUnlock()

	// TODO: This used to be a rudimentary check in the libp2p implementation whether the connecting node was known.
	//   In a gRPC implementation, this does not work any more, because the incoming connection uses a random port
	//   (and not the one the remote node is listening to). Use TLS and certificates to authenticate the remote node.
	_ = ok
	//if !ok {
	//	tr.logger.Log(logging.LevelWarn, "Received message from unknown peer. Stopping incoming connection",
	//		"remotePeer", peerAddr)
	//	return nil
	//}

	// Create a goroutine watching for the stop signal (closing tr.stop) and closing the connection on shutdown.
	// It is necessary for the case where the handler is blocked on tr.readAndProcessMessages and needs to shut down.
	connStopChan := make(chan struct{})
	defer close(connStopChan)
	go func() {
		select {
		case <-tr.stop:
			closeOnce.Do(closeFunc)
		case <-connStopChan:
			// Do nothing here.
			// This is only used to return from this goroutine if the connection handler stops by itself.
		}
	}()

	// Start reading and processig incoming messages.
	// This call blocks until an error occurs (e.g. the connection breaks) or the Transport is explicitly shut down.
	tr.readAndProcessMessages(srv, peerAddr)

	return nil
}

func (tr *Transport) getConnection(nodeID stdtypes.NodeID) (connection, error) {
	tr.connectionsLock.RLock()
	defer tr.connectionsLock.RUnlock()

	conn, ok := tr.connections[nodeID]
	if !ok {
		return nil, es.Errorf("no connection to node: %v", nodeID)
	}
	return conn, nil
}

// readAndProcessMessages Reads incoming data from the stream s, parses it to Mir messages,
// and writes them to the incomingMessages channel
func (tr *Transport) readAndProcessMessages(srv GrpcTransport_ListenServer, peerAddr string) {
	// Declare loop variables outside, since err is used also after the loop finishes.
	var err error
	var grpcMsg *GrpcMessage
	var destModule stdtypes.ModuleID

	// For each message received
	for grpcMsg, err = srv.Recv(); err == nil; grpcMsg, err = srv.Recv() {

		var rcvEvent stdtypes.Event
		switch msg := grpcMsg.Type.(type) {
		case *GrpcMessage_PbMsg:
			destModule = stdtypes.ModuleID(msg.PbMsg.DestModule)
			rcvEvent = transportpbevents.MessageReceived(
				destModule,
				stdtypes.NodeID(grpcMsg.Sender),
				messagepbtypes.MessageFromPb(msg.PbMsg),
			).Pb()
		case *GrpcMessage_RawMsg:
			destModule = stdtypes.ModuleID(msg.RawMsg.DestModule)
			rcvEvent = stdevents.NewMessageReceived(
				destModule,
				stdtypes.NodeID(grpcMsg.Sender),
				stdtypes.RawMessage(msg.RawMsg.Data),
			)
		}

		if tr.stats != nil {
			tr.stats.Received(proto.Size(grpcMsg), string(destModule.Top()))
		}

		select {
		case tr.incomingMessages <- stdtypes.ListOf(rcvEvent):
			// Write the message to the channel. This channel will be read by the user of the module.

		case <-srv.Context().Done():
			// If the connection closes before all its messages have been processed, ignore the unprocessed messages.
			tr.logger.Log(logging.LevelDebug, "Ignoring message, connection terminating.",
				"addr", peerAddr, err, err)
			break
		}
	}

	// Log error message produced on termination of the above loop.
	tr.logger.Log(logging.LevelInfo, "Connection terminated.", "addr", peerAddr, "err", err)
}
