package common

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleParams represents the state associated with a single instance of the checkpoint protocol
// (establishing a single stable checkpoint).
type ModuleParams struct {
	logging.Logger

	// The ID of the node executing this instance of the protocol.
	OwnID t.NodeID

	// Epoch to which this checkpoint belongs.
	// It is always the Epoch the checkpoint's associated sequence number (SeqNr) is part of.
	Epoch tt.EpochNr

	// Sequence number associated with this checkpoint protocol instance.
	// This checkpoint encompasses SeqNr sequence numbers,
	// i.e., SeqNr is the first sequence number *not* encompassed by this checkpoint.
	// One can imagine that the checkpoint represents the state of the system just before SeqNr,
	// i.e., "between" SeqNr-1 and SeqNr.
	SeqNr tt.SeqNr

	// The IDs of nodes to execute this instance of the checkpoint protocol.
	// Note that it is the Membership of Epoch e-1 that constructs the Membership for Epoch e.
	// (As the starting checkpoint for e is the "finishing" checkpoint for e-1.)
	Membership []t.NodeID

	// Time interval for repeated retransmission of checkpoint messages.
	ResendPeriod types.Duration
}

// EmptyContext type is used by DSL Origin-Result style event handlers in order to signal that no context is needed to be shared
type EmptyContext struct{}
