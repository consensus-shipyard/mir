/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	manet "github.com/multiformats/go-multiaddr/net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	mirnet "github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	transportpbtypes "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

const (
	// Maximum size of a gRPC message
	maxMessageSize = 1073741824
)

var _ mirnet.Transport = &Transport{}

// Transport represents a networking module that is based on gRPC.
// Each node's networking module contains one gRPC server, to which other nodes' modules connect.
// The type of gRPC connection is multi-request-single-response, where each module contains
// one instance of a gRPC client per node.
// A message to a node is sent as request to that node's gRPC server.
type Transport struct {
	UnimplementedGrpcTransportServer

	// The ID of the node that uses this networking module.
	ownID t.NodeID

	// The address of the node.
	ownAddr t.NodeAddress

	// Channel to which all incoming messages are written.
	// This channel is also returned by the ReceiveChan() method.
	incomingMessages chan *events.EventList

	// For each node ID, stores a gRPC message sink, calling the Send() method of which sends a message to that node.
	clients map[t.NodeID]GrpcTransport_ListenClient

	// For each node ID, stores a gRPC connection to that node.
	conns map[t.NodeID]*grpc.ClientConn

	// The gRPC server used by this networking module.
	grpcServer *grpc.Server

	// Error returned from the grpcServer.Serve() call (see Start() method).
	grpcServerError error

	// Logger use for all logging events of this GrpcTransport
	logger logging.Logger
}

// NewTransport returns a pointer to a new initialized GrpcTransport networking module.
// The membership parameter must represent the complete static membership of the system.
// It maps the node ID of each node in the system to
// a string representation of its network address with the format "IPAddress:port".
// The ownId parameter is the ID of the node that will use the returned networking module.
// The returned GrpcTransport is not yet running (able to receive messages),
// nor is it connected to any nodes (able to send messages).
// This needs to be done explicitly by calling the respective Start() and Connect() methods.
func NewTransport(id t.NodeID, addr t.NodeAddress, l logging.Logger) (*Transport, error) {

	// If no logger was given, only write errors to the console.
	if l == nil {
		l = logging.ConsoleErrorLogger
	}

	return &Transport{
		ownID:            id,
		ownAddr:          addr,
		incomingMessages: make(chan *events.EventList),
		clients:          make(map[t.NodeID]GrpcTransport_ListenClient),
		conns:            make(map[t.NodeID]*grpc.ClientConn),
		logger:           l,
	}, nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (gt *Transport) ImplementsModule() {}

func (gt *Transport) EventsOut() <-chan *events.EventList {
	return gt.incomingMessages
}

func (gt *Transport) ApplyEvents(
	ctx context.Context,
	eventList *events.EventList,
) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		switch e := event.Type.(type) {
		case *eventpb.Event_Init:
			// no actions on init
		case *eventpb.Event_Transport:
			switch e := transportpbtypes.EventFromPb(e.Transport).Type.(type) {
			case *transportpbtypes.Event_SendMessage:
				for _, destID := range e.SendMessage.Destinations {
					if destID == gt.ownID {
						// Send message to myself bypassing the network.
						// The sending must be done in its own goroutine in case writing to gt.incomingMessages blocks.
						// (Processing of input events must be non-blocking.)
						receivedEvent := transportpbevents.MessageReceived(
							e.SendMessage.Msg.DestModule,
							gt.ownID,
							e.SendMessage.Msg,
						)
						go func() {
							select {
							case gt.incomingMessages <- events.ListOf(receivedEvent.Pb()):
							case <-ctx.Done():
							}
						}()
					} else {
						// Send message to another node.
						if err := gt.Send(destID, e.SendMessage.Msg.Pb()); err != nil {
							// TODO: This violates the non-blocking operation of ApplyEvents method. Fix it.
							gt.logger.Log(logging.LevelWarn, "failed to send a message", "err", err)
						}
					}
				}
			default:
				return fmt.Errorf("unexpected type of transport event: %T", e)
			}
		default:
			return fmt.Errorf("unexpected type of Net event: %T", event.Type)
		}
	}

	return nil
}

// Send sends msg to the node with ID dest.
// Concurrent calls to Send are not (yet? TODO) supported.
func (gt *Transport) Send(dest t.NodeID, msg *messagepb.Message) error {
	return gt.clients[dest].Send(&GrpcMessage{Sender: gt.ownID.Pb(), Msg: msg})
}

// Listen implements the gRPC Listen service (multi-request-single-response).
// It receives messages from the gRPC client running on the other node
// and writes them to a channel that the user can access through ReceiveChan().
// This function is called by the gRPC system on every new connection
// from another node's Net module's gRPC client.
func (gt *Transport) Listen(srv GrpcTransport_ListenServer) error {

	// Print address of incoming connection.
	p, ok := peer.FromContext(srv.Context())
	if ok {
		gt.logger.Log(logging.LevelDebug, fmt.Sprintf("Incoming connection from %s", p.Addr.String()))
	} else {
		return fmt.Errorf("failed to get grpc peer info from context")
	}

	// Declare loop variables outside, since err is used also after the loop finishes.
	var err error
	var grpcMsg *GrpcMessage

	// For each message received
	for grpcMsg, err = srv.Recv(); err == nil; grpcMsg, err = srv.Recv() {
		select {
		case gt.incomingMessages <- events.ListOf(
			transportpbevents.MessageReceived(
				t.ModuleID(grpcMsg.Msg.DestModule),
				t.NodeID(grpcMsg.Sender),
				messagepbtypes.MessageFromPb(grpcMsg.Msg),
			).Pb(),
		):
			// Write the message to the channel. This channel will be read by the user of the module.

		case <-srv.Context().Done():
			// If the connection closes before all its messages have been processed, ignore the unprocessed messages.
			gt.logger.Log(logging.LevelDebug, "Ignoring message, connection terminating.",
				"addr", p.Addr.String(), err, err)
			break
		}
	}

	// Log error message produced on termination of the above loop.
	gt.logger.Log(logging.LevelInfo, "Connection terminated.", "addr", p.Addr.String(), err, err)

	// Send gRPC response message and close connection.
	return srv.SendAndClose(&ByeBye{})
}

// Start starts the networking module by initializing and starting the internal gRPC server,
// listening on the port determined by the membership and own ID.
// Before ths method is called, no other GrpcTransports can connect to this one.
func (gt *Transport) Start() error {
	// Obtain net.Dial compatible address.
	_, dialAddr, err := manet.DialArgs(gt.ownAddr)
	if err != nil {
		return fmt.Errorf("failed to obtain Dial address: %w", err)
	}
	// Obtain own port number from membership.
	_, ownPort, err := net.SplitHostPort(dialAddr)
	if err != nil {
		return err
	}

	gt.logger.Log(logging.LevelInfo, fmt.Sprintf("Listening for connections on port %s", ownPort))

	// Create a gRPC server and assign it the logic of this module.
	gt.grpcServer = grpc.NewServer()
	RegisterGrpcTransportServer(gt.grpcServer, gt)

	// Start listening on the network
	conn, err := net.Listen("tcp", ":"+ownPort)
	if err != nil {
		return fmt.Errorf("failed to listen for connections on port %s: %w", ownPort, err)
	}

	// Start the gRPC server in a separate goroutine.
	// When the server stops, it will write its exit error into gt.grpcServerError.
	go func() {
		gt.grpcServerError = gt.grpcServer.Serve(conn)
	}()

	// If we got all the way here, no error occurred.
	return nil
}

// Stop closes all open connections to other nodes and stops the own gRPC server
// (preventing further incoming connections).
// After Stop() returns, the error returned by the gRPC server's Serve() call
// can be obtained through the ServerError() method.
func (gt *Transport) Stop() {
	gt.logger.Log(logging.LevelDebug, "Stopping gRPC transport.")
	defer gt.logger.Log(logging.LevelDebug, "gRPC transport stopped.")

	// Close connections to other nodes.
	for id, client := range gt.clients {
		if client == nil {
			continue
		}
		gt.logger.Log(logging.LevelDebug, "Closing gRPC client", "to", id)

		if err := client.CloseSend(); err != nil {
			gt.logger.Log(logging.LevelWarn, fmt.Sprintf("Could not close gRPC client to node %v: %v", id, err))
			continue
		}

		gt.logger.Log(logging.LevelDebug, "Closed gRPC client", "to", id)
	}

	// Close connections to other nodes.
	for id, conn := range gt.conns {
		if conn == nil {
			continue
		}
		gt.logger.Log(logging.LevelDebug, "Closing connection", "to", id)

		if err := conn.Close(); err != nil {
			gt.logger.Log(logging.LevelWarn, fmt.Sprintf("Could not close connection to node %v: %v", id, err))
			continue
		}

		gt.logger.Log(logging.LevelDebug, "Closed connection", "to", id)
	}

	// Stop own gRPC server.
	gt.logger.Log(logging.LevelDebug, "Stopping gRPC server")
	gt.grpcServer.GracefulStop()
	gt.logger.Log(logging.LevelDebug, "gRPC server stopped")
}

// ServerError returns the error returned by the gRPC server's Serve() call.
// ServerError() must not be called before the GrpcTransport is stopped and its Stop() method has returned.
func (gt *Transport) ServerError() error {
	return gt.grpcServerError
}

func (gt *Transport) CloseOldConnections(newNodes map[t.NodeID]t.NodeAddress) {
	for id, client := range gt.clients {
		if client == nil {
			continue
		}

		// Close an old connection to a node if we don't need to connect to this node further.
		if _, newConn := newNodes[id]; !newConn {
			gt.logger.Log(logging.LevelDebug, "Closing old connection", "to", id)

			if err := client.CloseSend(); err != nil {
				gt.logger.Log(logging.LevelError, fmt.Sprintf("Could not close old client to node %v: %v", id, err))
				continue
			}

			delete(gt.clients, id)

			gt.logger.Log(logging.LevelDebug, "Closed old client", "to", id)
		}
	}
	for id, conn := range gt.conns {
		if conn == nil {
			continue
		}

		// Close an old connection to a node if we don't need to connect to this node further.
		if _, newConn := newNodes[id]; !newConn {
			gt.logger.Log(logging.LevelDebug, "Closing old connection", "to", id)

			if err := conn.Close(); err != nil {
				gt.logger.Log(logging.LevelError, fmt.Sprintf("Could not close old connection to node %v: %v", id, err))
				continue
			}

			delete(gt.conns, id)

			gt.logger.Log(logging.LevelDebug, "Closed old connection", "to", id)
		}
	}
}

// Connect establishes (in parallel) network connections to all nodes according to the membership table.
// The other nodes' GrpcTransport modules must be running.
// Only after Connect() returns, sending messages over this GrpcTransport is possible.
// TODO: Deal with errors, e.g. when the connection times out (make sure the RPC call in connectToNode() has a timeout).
func (gt *Transport) Connect(nodes map[t.NodeID]t.NodeAddress) {

	// Initialize wait group used by the connecting goroutines
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))

	// Synchronizes concurrent access to connections.
	lock := sync.Mutex{}

	// For each node in the membership
	for nodeID, nodeAddr := range nodes {

		// Get net.Dial compatible address.
		_, dialAddr, err := manet.DialArgs(nodeAddr)
		if err != nil {
			wg.Done()
			continue
		}

		// Launch a goroutine that connects to the node.
		go func(id t.NodeID, addr string) {
			defer wg.Done()

			// Create and store connection
			conn, client, err := gt.connectToNode(addr) // May take long time, execute before acquiring the lock.
			lock.Lock()
			gt.clients[id] = client
			gt.conns[id] = conn
			lock.Unlock()

			// Print debug info.
			if err != nil {
				gt.logger.Log(logging.LevelError,
					fmt.Sprintf("Failed to connect to node %v (%s): %v", id, addr, err))
			} else {
				gt.logger.Log(logging.LevelDebug, fmt.Sprintf("Node %v (%s) connected.", id, addr))
			}

		}(nodeID, dialAddr)
	}

	// Wait for connecting goroutines to finish.
	wg.Wait()
}

func (gt *Transport) WaitFor(_ int) {
	// TODO: We return immediately here, as the Connect() function already waits for all connections to be established.
	//       This is not right and should be done as in the libp2p transport.
}

// Establishes a connection to a single node at address addrString.
func (gt *Transport) connectToNode(addrString string) (*grpc.ClientConn, GrpcTransport_ListenClient, error) {

	gt.logger.Log(logging.LevelDebug, fmt.Sprintf("Connecting to node: %s", addrString))

	// Set general gRPC dial options.
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize), grpc.MaxCallSendMsgSize(maxMessageSize)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Set up a gRPC connection.
	conn, err := grpc.Dial(addrString, dialOpts...)
	if err != nil {
		return nil, nil, err
	}

	// Register client stub.
	client := NewGrpcTransportClient(conn)

	// Remotely invoke the Listen function on the other node's gRPC server.
	// As this is "stream of requests"-type RPC, it returns a message sink.
	msgSink, err := client.Listen(context.Background())
	if err != nil {
		if cerr := conn.Close(); cerr != nil {
			gt.logger.Log(logging.LevelWarn, fmt.Sprintf("Failed to close connection: %v", cerr))
		}
		return nil, nil, err
	}

	// Return the message sink connected to the node.
	return conn, msgSink, nil
}
