package net

import (
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type Transport interface {
	modules.ActiveModule

	// Start starts the networking module by initializing and starting the corresponding network services.
	Start() error

	// Stop closes all open connections to other nodes and stops the network services.
	Stop()

	// Send sends msg to the node with ID dest.
	// TODO: Remove this method from the interface definition. Sending is invoked by event processing, not externally.
	Send(dest t.NodeID, msg *messagepb.Message) error

	// Connect initiates the establishing of network connections to the provided nodes.
	// When Connect returns, the connections might not yet have been established though (see WaitFor).
	Connect(nodes *trantorpbtypes.Membership)

	// WaitFor waits until at least n connections (including the potentially virtual connection to self)
	// have been established and returns nil.
	// If the networking module is stopped while WaitFor is invoked, WaitFor returns a non-nil error.
	WaitFor(n int) error

	// CloseOldConnections closes connections to the nodes that don't needed.
	CloseOldConnections(newNodes *trantorpbtypes.Membership)
}
