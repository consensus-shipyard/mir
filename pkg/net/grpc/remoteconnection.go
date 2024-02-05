package grpc

import (
	"context"
	"fmt"
	"sync"

	es "github.com/go-errors/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	// Maximum size of a gRPC message
	maxMessageSize = 1073741824
)

type remoteConnection struct {
	params Params

	// The address of the target node.
	// It must already be in a format accepted by grpc.DialContext
	// (can be produced by manet.DialArgs from a Multiaddress)
	addr string

	msgBuffer chan *GrpcMessage
	stop      chan struct{}
	done      chan struct{}

	clientConn    *grpc.ClientConn
	msgSink       GrpcTransport_ListenClient
	connectedCond *sync.Cond

	stats  net.Stats
	logger logging.Logger
}

func newRemoteConnection(
	params Params,
	addr string,
	logger logging.Logger,
	stats net.Stats,
) (*remoteConnection, error) {
	conn := &remoteConnection{
		params:        params,
		addr:          addr,
		logger:        logger,
		stats:         stats,
		clientConn:    nil,
		msgSink:       nil,
		msgBuffer:     make(chan *GrpcMessage, params.ConnectionBufferSize),
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
		connectedCond: sync.NewCond(&sync.Mutex{}),
	}
	go conn.process()
	return conn, nil
}

// Address returns the network address of the other side of this connection.
func (conn *remoteConnection) Address() string {
	return conn.addr
}

// Send makes a non-blocking attempt to send a message to this connection.
// Send might use internal buffering. Thus, even if it returns nil,
// the message might not have yet been sent to the network.
func (conn *remoteConnection) Send(msg *GrpcMessage) error {

	select {
	case conn.msgBuffer <- msg:
		return nil
	default:
		return es.Errorf("send buffer full (" + conn.addr + ")")
	}
}

// Close closes the connection. No data will be sent to the underlying network stream after Close returns.
func (conn *remoteConnection) Close() {

	// Do nothing if connection already has been closed.
	select {
	case <-conn.stop:
		return
	default:
	}

	// Stop processing and wait until it finishes.
	conn.connectedCond.L.Lock()
	close(conn.stop)
	conn.connectedCond.Broadcast()
	conn.connectedCond.L.Unlock()
	<-conn.done
}

// Wait returns an error channel and a cancel function.
// The channel will be closed without any value being written to it
// when the underlying network stream has been established.
// Waiting is aborted when the cancel function is called or when the connection is closed.
// In both cases, an error is written in the returned channel.
func (conn *remoteConnection) Wait() (chan error, func()) {

	// The channel to be returned.
	result := make(chan error, 1)

	// This flag is set by the returned abort function.
	// It makes the goroutine waiting for the connection return an error.
	abort := false

	go func() {
		conn.connectedCond.L.Lock()
		defer conn.connectedCond.L.Unlock()

		// Wait while
		for conn.clientConn == nil && !abort {
			// the connection has not yet been established and the caller of Wait has not called the abort function
			select {
			case <-conn.stop:
				// and the connection is not closing.
				result <- es.Errorf("connection closed")
				return
			default:
				conn.connectedCond.Wait()
			}
		}

		// If the waiting was aborted, output an error, otherwise exit successfully (by closing the channel).
		if abort {
			result <- es.Errorf("waiting aborted")
		} else {
			close(result)
		}
	}()

	// Return the result channel and the cancel function.
	return result, func() {
		conn.connectedCond.L.Lock()
		abort = true
		conn.connectedCond.Broadcast()
		conn.connectedCond.L.Unlock()
	}
}

// connect establishes the underlying network stream.
// In case of a network failure, connect keeps retrying until it succeeds (returning nil)
// or until the connection is closed (returning a non-nil error).
// connect blocks until the stream is created or Close is called.
func (conn *remoteConnection) connect() error {

	// Create a context that will be canceled when the connection is closed.
	// It will be used for aborting opening the stream.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // This is necessary for stopping the goroutine below in case of a successful connection.
	go func() {
		select {
		case <-ctx.Done():
			// Garbage-collect this goroutine when the context it is supposed to cancel is canceled by someone else.
		case <-conn.stop:
			// Cancel the new stream creation if connection is closed.
			cancel()
		}
	}()

	// Retry connecting until we succeed.
	for conn.clientConn == nil {

		select {
		case <-conn.stop:
			// Stop connection attempts if connection is closing.
			return es.Errorf("context canceled")
		default:

			// Try connecting to the peer.
			if err := conn.tryConnecting(); err != nil {
				return err
			}
		}
	}

	return nil
}

// tryConnecting makes a single attempt to open the underlying network stream.
// It blocks until it succeeds or, if it fails, at least for ReconnectionPeriod (unless the connection is closing).
// tryConnecting returns nil on success or if it is meaningful to retry connecting (i.e., call tryConnecting again).
// It only returns a non-nil error if it is not meaningful to retry.
func (conn *remoteConnection) tryConnecting() error {

	conn.logger.Log(logging.LevelDebug, fmt.Sprintf("Connecting to node: %s", conn.addr))

	timeoutCtx, cancel := context.WithTimeout(context.Background(), conn.params.ReconnectionPeriod)
	defer cancel()

	// Set general gRPC dial options.
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize), grpc.MaxCallSendMsgSize(maxMessageSize)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Set up a gRPC connection.
	clientConn, err := grpc.DialContext(timeoutCtx, conn.addr, dialOpts...)
	if err != nil {
		conn.logger.Log(logging.LevelWarn, "Failed dialing.", "addr", conn.addr, "err", err)
		return nil // Returning indicates that another attempt to connect should be made.
	}

	// Register client stub.
	client := NewGrpcTransportClient(clientConn)

	// Remotely invoke the Listen function on the other node's gRPC server.
	// As this is "stream of requests"-type RPC, it returns a message sink.
	// We don't use timeoutCtx here, as we need the msgSink to survive even after the timeout.
	msgSink, err := client.Listen(context.Background())
	if err != nil {
		if cerr := clientConn.Close(); cerr != nil {
			conn.logger.Log(logging.LevelWarn, fmt.Sprintf("Failed to close connection: %v", cerr))
		}
		return nil // Returning indicates that another attempt to connect should be made.
	}

	// If connecting failed, wait until the reconnection timeout expires before returning.
	// (tryConnecting is likely to be called again immediately if it fails.)
	if err != nil {
		select {
		case <-conn.stop:
			return es.Errorf("connection closing")
		case <-timeoutCtx.Done():
			conn.logger.Log(logging.LevelWarn, "Failed connecting.", "err", err)
			return nil
		}
	}

	// If connecting succeeded, save the new stream
	// and notify any goroutines waiting for the connection establishment (Wait method).
	conn.connectedCond.L.Lock()
	conn.clientConn = clientConn
	conn.msgSink = msgSink
	conn.connectedCond.Broadcast()
	conn.connectedCond.L.Unlock()
	return nil
}

// process is the main processing loop.
// It keeps reading the input data buffer and writes its contents to the network.
// It automatically creates the underlying network stream and re-establishes it as needed, in case it is dropped.
func (conn *remoteConnection) process() {
	// In case there is a panic in the main processing loop, log an error message.
	// (Otherwise, since this function is run as a goroutine, panicking would be completely silent.)
	defer func() {
		if r := recover(); r != nil {
			err := es.New(r)
			conn.logger.Log(logging.LevelError, "Remote connection panicked.", "cause", r, "stack", err.ErrorStack())
		}
	}()

	// When processing finishes, close the underlying stream and signal to the Stop method that it can return.
	// Note that the defer order is thus inverted.
	defer close(conn.done)
	defer conn.closeClientConn()

	// Message to be sent to the connection.
	// If nil, a new message from conn.msgBuffer will be read, and stored here.
	var grpcMessage *GrpcMessage

	for {
		// The processing loop runs indefinitely (until interrupted by explicitly returning).
		// One iteration corresponds to one attempt of sending a message.

		// Check if connection is being closed.
		// This is necessary for not getting stuck trying to send an unsent message
		// (failing all the time and retrying forever).
		select {
		case <-conn.stop:
			return
		default:
		}

		// Create a network connection if there is none.
		if conn.clientConn == nil {
			if err := conn.connect(); err != nil {
				// Unless the connection is closing, connect() will keep retrying to connect indefinitely.
				// Thus, if it returns an error, it means that there is no point in continuing the processing.
				conn.logger.Log(logging.LevelWarn, "Gave up connecting", "err", err)
				return
			}
		}

		// Get the next message if there is no pending unsent message.
		if grpcMessage == nil {
			select {
			case <-conn.stop:
				return
			case grpcMessage = <-conn.msgBuffer:
			}
		}

		// Write the encoded data to the network stream.
		if err := conn.sendGrpcMessage(grpcMessage); err != nil {
			// If writing fails, close the stream, such that a new one will be re-established in the next iteration.
			conn.logger.Log(logging.LevelWarn, "Failed sending message.", "err", err)
			conn.closeClientConn()
		} else {
			// On success, clear the pending message (that has just been sent)
			// so a new one can be read from the msbBuffer on the next iteration.
			grpcMessage = nil
		}
	}
}

// sendGrpcMessage writes data to the underlying network stream.
// It blocks until all data is written, the connection closes, or an error occurs.
// In the first case, sendGrpcMessage returns nil. Otherwise, it returns the corresponding error.
func (conn *remoteConnection) sendGrpcMessage(grpcMessage *GrpcMessage) error {

	// Label to associate the data with. Only relevant for recording statistics.
	var statsLabel string
	switch m := grpcMessage.Type.(type) {
	case *GrpcMessage_PbMsg:
		statsLabel = string(stdtypes.ModuleID(m.PbMsg.DestModule).Top())
	case *GrpcMessage_RawMsg:
		statsLabel = string(stdtypes.ModuleID(m.RawMsg.DestModule).Top())
	default:
		panic(es.Errorf("unsupported message type: %T", grpcMessage.Type))
	}

	if err := conn.msgSink.Send(grpcMessage); err != nil {
		return es.Errorf("failed sending grpc message: %w", err)
	}

	if conn.stats != nil {
		conn.stats.Sent(proto.Size(grpcMessage), statsLabel)
	}

	return nil
}

// closeClientConn closes the underlying network connection if it is open.
func (conn *remoteConnection) closeClientConn() {

	if conn.clientConn != nil {
		if err := conn.msgSink.CloseSend(); err != nil {
			conn.logger.Log(logging.LevelWarn, "Failed closing grpc client.", "err", err)
		}
		if err := conn.clientConn.Close(); err != nil {
			conn.logger.Log(logging.LevelWarn, "Failed closing grpc client connection.", "err", err)
		}

		// conn.clientConn == nil is used as a condition in the Wait method and thus needs to be guarded by the lock.
		// Also, conn.clientConn and conn.msgSink are expected both to be nil or both to be non-nil.
		conn.connectedCond.L.Lock()
		conn.clientConn = nil
		conn.msgSink = nil
		conn.connectedCond.L.Unlock()
	}
}
