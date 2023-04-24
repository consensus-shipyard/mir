package net

import (
	"github.com/filecoin-project/mir/pkg/modules"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type Transport interface {
	modules.ActiveModule

	// Start starts the networking module by initializing and starting the corresponding network services.
	Start() error

	// Stop closes all open connections to other nodes and stops the network services.
	Stop()

	// Send sends msg to the node with ID dest.
	Send(dest t.NodeID, msg *messagepb.Message) error

	// Connect initiates the establishing of network connections to the provided nodes.
	// When Connect returns, the connections might not yet have been established though (see WaitFor).
	Connect(nodes *commonpbtypes.Membership)

	// WaitFor waits until at least n connections (including the potentially virtual connection to self)
	// have been established and returns nil.
	// If the networking module is stopped while WaitFor is invoked, WaitFor returns a non-nil error.
	WaitFor(n int) error

	// CloseOldConnections closes connections to the nodes that don't needed.
	CloseOldConnections(newNodes *commonpbtypes.Membership)
}
