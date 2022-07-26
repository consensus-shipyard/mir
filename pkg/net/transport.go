package net

import (
	"context"

	"github.com/filecoin-project/mir/pkg/modules"
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

	// Connect establishes (in parallel) network connections to the provided nodes.
	Connect(ctx context.Context, nodes map[t.NodeID]t.NodeAddress)
}
