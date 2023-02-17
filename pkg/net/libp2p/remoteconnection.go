package libp2p

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-yamux/v3"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type remoteConnection struct {
	params        Params
	ownID         t.NodeID
	addrInfo      *peer.AddrInfo
	logger        logging.Logger
	host          host.Host
	stream        network.Stream
	msgBuffer     chan *messagepb.Message
	stop          chan struct{}
	done          chan struct{}
	connectedCond *sync.Cond
}

func newRemoteConnection(
	params Params,
	ownID t.NodeID,
	addr t.NodeAddress,
	h host.Host,
	logger logging.Logger,
) (*remoteConnection, error) {
	addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address: %w", err)
	}
	conn := &remoteConnection{
		params:        params,
		ownID:         ownID,
		addrInfo:      addrInfo,
		logger:        logger,
		host:          h,
		stream:        nil,
		msgBuffer:     make(chan *messagepb.Message, params.ConnectionBufferSize),
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
		connectedCond: sync.NewCond(&sync.Mutex{}),
	}
	go conn.process()
	return conn, nil
}

// PeerID returns the libp2p peer ID of the other side of this connection.
func (conn *remoteConnection) PeerID() peer.ID {
	return conn.addrInfo.ID
}

// Send makes a non-blocking attempt to send a message to this connection.
// Send might use internal buffering. Thus, even if it returns nil,
// the message might not have yet been sent to the network.
func (conn *remoteConnection) Send(msg *messagepb.Message) error {

	select {
	case conn.msgBuffer <- msg:
		return nil
	default:
		return fmt.Errorf("send buffer full")
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
		for conn.stream == nil && !abort {
			// the connection has not yet been established and the caller of Wait has not called the abort function
			select {
			case <-conn.stop:
				// and the connection is not closing.
				result <- fmt.Errorf("connection closed")
				return
			default:
				conn.connectedCond.Wait()
			}
		}

		// If the waiting was aborted, output an error, otherwise exit successfully (by closing the channel).
		if abort {
			result <- fmt.Errorf("waiting aborted")
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
	for conn.stream == nil {

		select {
		case <-conn.stop:
			// Stop connection attempts if connection is closing.
			return fmt.Errorf("context canceled")
		default:

			// Try connecting to the peer.
			if err := conn.tryConnecting(ctx); err != nil {
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
func (conn *remoteConnection) tryConnecting(ctx context.Context) error {

	// Try creating a network stream.
	conn.host.Peerstore().AddAddrs(conn.addrInfo.ID, conn.addrInfo.Addrs, conn.params.ConnectionTTL)
	stream, err := conn.host.NewStream(ctx, conn.addrInfo.ID, conn.params.ProtocolID)

	// If connecting failed, wait a moment before returning.
	// (tryConnecting is likely to be called again immediately if it fails.)
	after := time.NewTimer(conn.params.ReconnectionPeriod)
	defer after.Stop()

	if err != nil {
		select {
		case <-conn.stop:
			return fmt.Errorf("context canceled")
		case <-after.C:
			conn.logger.Log(logging.LevelWarn, "Failed connecting.", "err", err)
			return nil
		}
	}

	// If connecting succeeded, save the new stream
	// and notify any goroutines waiting for the connection establishment (Wait method).
	conn.connectedCond.L.Lock()
	conn.stream = stream
	conn.connectedCond.Broadcast()
	conn.connectedCond.L.Unlock()
	return nil
}

// process is the main processing loop.
// It keeps reading the input data buffer and writes its contents to the network.
// It automatically creates the underlying network stream and re-establishes it as needed, in case it is dropped.
func (conn *remoteConnection) process() {
	// When processing finishes, close the underlying stream and signal to the Stop method that it can return.
	// Note that the defer order is thus inverted.
	defer close(conn.done)
	defer conn.closeStream()

	for {
		// The processing loop runs indefinitely (until interrupted by explicitly returning).
		// One iteration corresponds to sending one message.

		// Create a network connection if there is none.
		if conn.stream == nil {
			if err := conn.connect(); err != nil {
				// Unless the connection is closing, connect() will keep retrying to connect indefinitely.
				// Thus, if it returns an error, it means that there is no point in continuing the processing.
				conn.logger.Log(logging.LevelWarn, "Gave up connecting", "err", err)
				return
			}
		}

		// Get the next message and write it to the output stream (unless connection is closing).
		select {
		case <-conn.stop:
			return
		case msg := <-conn.msgBuffer:

			// Encode message to a byte slice.
			data, err := encodeMessage(msg, conn.ownID)
			if err != nil {
				conn.logger.Log(logging.LevelError, "Could not encode message. Disconnecting.", "err", err)
				return
			}

			// Write the encoded data to the network stream.
			// If writing fails, close the stream, such that a new one will be re-established in the next iteration.
			if err = conn.writeDataToStream(data); err != nil {
				conn.logger.Log(logging.LevelWarn, "Failed sending data.", "err", err)
				conn.closeStream()
			}
		}
	}
}

// writeDataToStream writes data to the underlying network stream.
// It blocks until all data is written, the connection closes, or an error occurs.
// In the first case, writeDataToStream returns nil. Otherwise, it returns the corresponding error.
func (conn *remoteConnection) writeDataToStream(data []byte) error {

	// Retry sending data until:
	// - all data is sent, or
	// - the connection closes, or
	// - an error occurs.
	for {

		// Set a timeout for the data to be written, so the conn.stream.Write call does not block forever.
		// This is required so that we can periodically check the conn.stop channel.
		if err := conn.stream.SetWriteDeadline(time.Now().Add(conn.params.StreamWriteTimeout)); err != nil {
			return fmt.Errorf("could not set stream write deadline")
		}

		// Try writing data to the underlying network stream.
		bytesWritten, err := conn.stream.Write(data)

		if err == nil {
			// If all data was successfully written, return.

			return nil

		} else if errors.Is(err, yamux.ErrTimeout) {
			// If a timeout occurred, check if the connection has not been closed in the meantime.
			// If the connection is still open, retry sending the rest of the data in the next iteration.

			select {
			case <-conn.stop:
				return fmt.Errorf("connection closing")
			default:
				data = data[bytesWritten:]
			}

		} else {
			// If any other error occurred, just return it.

			return fmt.Errorf("failed sending data: %w", err)

		}
	}
}

// closeStream closes the underlying network stream if it is open.
func (conn *remoteConnection) closeStream() {

	if conn.stream != nil {
		if err := conn.stream.Close(); err != nil {
			conn.logger.Log(logging.LevelWarn, "Failed closing stream.", "err", err)
		}

		// conn.stream == nil is used as a condition in the Wait method and thus needs to be guarded by the lock.
		conn.connectedCond.L.Lock()
		conn.stream = nil
		conn.connectedCond.L.Unlock()
	}
}

func encodeMessage(msg *messagepb.Message, nodeID t.NodeID) ([]byte, error) {
	p, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	tm := TransportMessage{nodeID.Pb(), p}
	buf := new(bytes.Buffer)
	if err = tm.MarshalCBOR(buf); err != nil {
		return nil, fmt.Errorf("failed to CBOR marshal message: %w", err)
	}
	return buf.Bytes(), nil
}
