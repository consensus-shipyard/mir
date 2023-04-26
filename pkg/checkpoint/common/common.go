package common

import (
	"github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleParams represents the state associated with a single instance of the checkpoint protocol
// (establishing a single stable checkpoint).
type ModuleParams struct {
	logging.Logger

	// IDs of modules the checkpoint tracker interacts with.
	// TODO: Eventually put the checkpoint tracker in a separate package and create its own ModuleConfig type.
	ModuleConfig *ModuleConfig

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

	// State snapshot associated with this checkpoint.
	StateSnapshot *trantorpbtypes.StateSnapshot

	// Hash of the state snapshot data associated with this checkpoint.
	StateSnapshotHash []byte

	// Set of (potentially invalid) nodes' Signatures.
	Signatures map[t.NodeID][]byte

	// Set of nodes from which a valid Checkpoint messages has been received.
	Confirmations map[t.NodeID]struct{}

	// Set of Checkpoint messages that were received ahead of time.
	PendingMessages map[t.NodeID]*checkpointpbtypes.Checkpoint

	// Time interval for repeated retransmission of checkpoint messages.
	ResendPeriod types.Duration

	// Flag ensuring that the stable checkpoint is only Announced once.
	// Set to true when announcing a stable checkpoint for the first time.
	// When true, stable checkpoints are not Announced anymore.
	Announced bool
}

func (p *ModuleParams) SnapshotReady() bool {
	return p.StateSnapshot.AppData != nil &&
		p.StateSnapshot.EpochData.ClientProgress != nil
}

func (p *ModuleParams) Stable() bool {
	return p.SnapshotReady() && len(p.Confirmations) >= config.StrongQuorum(len(p.Membership))
}
