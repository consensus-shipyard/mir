package libp2p

import (
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/filecoin-project/mir/pkg/pb/messagepb"
)

// connection represents a connection to a (local or remote) peer.
type connection interface {

	// PeerID returns the libp2p peer ID of the other side of this connection.
	PeerID() peer.ID

	// Send makes a non-blocking attempt to send a message to this connection.
	// Send might use internal buffering. Thus, even if it returns nil,
	// the message might not have yet been physically sent.
	Send(message *messagepb.Message) error

	// Close closes the connection. No data will be sent to the underlying stream after Close returns.
	Close()

	// Wait returns an error channel and a cancel function.
	// The channel will be closed without any value being written to it
	// when the underlying network stream has been established.
	// Waiting is aborted when the cancel function is called or when the connection is closed.
	// In both cases, an error is written in the returned channel.
	Wait() (chan error, func())
}
