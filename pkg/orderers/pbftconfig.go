package orderers

import (
	"time"

	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// PBFTConfig holds PBFT-specific configuration parameters used by a concrete instance of PBFT.
// They are mostly inherited from the ISS configuration at the time of creating the PBFT instance.
type PBFTConfig struct {

	// The IDs of all nodes that execute this instance of the protocol.
	// Must not be empty.
	Membership []t.NodeID

	// The maximum time duration between two proposals of new certificatees during normal operation.
	// This parameter caps the waiting time in order to bound latency.
	// When MaxProposeDelay has elapsed since the last proposal,
	// the protocol tries to propose a new availability certificate.
	// Must not be negative.
	MaxProposeDelay time.Duration

	// When a node has committed all certificates in a segment, it will periodically send the Done message
	// in intervals of DoneResendPeriod.
	DoneResendPeriod time.Duration

	// After a node learns about a quorum of other nodes finishing a segment,
	// it waits for CatchUpDelay before requesting missing committed certificates from other nodes.
	CatchUpDelay time.Duration

	// Maximal number of bytes used for message backlogging buffers
	// (only message payloads are counted towards MsgBufCapacity).
	// Same as ModuleParams.MsgBufCapacity, but used only for one instance of PBFT.
	// Must not be negative.
	MsgBufCapacity int

	// Per-sequence-number view change timeout for view 0.
	// If no certificate is delivered by a PBFT instance within this timeout, the node triggers a view change.
	// With each new view, the timeout doubles (without changing this value)
	ViewChangeSNTimeout time.Duration

	// View change timeout for view 0 for the whole segment.
	// If not all certificates of the associated segment are delivered by a PBFT instance within this timeout,
	// the node triggers a view change.
	// With each new view, the timeout doubles (without changing this value)
	ViewChangeSegmentTimeout time.Duration

	// Time period between resending a ViewChange message.
	// ViewChange messages need to be resent periodically to preserve liveness.
	// Otherwise, the system could get stuck if a ViewChange message is dropped by the network.
	ViewChangeResendPeriod time.Duration

	// The current epoch number
	epochNr tt.EpochNr
}
