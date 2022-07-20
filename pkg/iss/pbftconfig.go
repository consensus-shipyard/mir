package iss

import (
	"time"

	t "github.com/filecoin-project/mir/pkg/types"
)

// PBFTConfig holds PBFT-specific configuration parameters used by a concrete instance of PBFT.
// They are mostly inherited from the ISS configuration at the time of creating the PBFT instance.
type PBFTConfig struct {

	// The IDs of all nodes that execute this instance of the protocol.
	// Must not be empty.
	Membership []t.NodeID

	// The maximum time duration between two proposals of new batches during normal operation.
	// This parameter caps the waiting time in order to bound latency.
	// When MaxProposeDelay has elapsed since the last proposal,
	// the protocol tries to propose a new request batch, even if the batch is not full (or even completely empty).
	// Must not be negative.
	MaxProposeDelay time.Duration

	// When a node has committed all batches in a segment, it will periodically send the Done message
	// in intervals of DoneResendPeriod.
	DoneResendPeriod time.Duration

	// After a node learns about a quorum of other nodes finishing a segment,
	// it waits for CatchUpDelay before requesting missing committed batches from other nodes.
	CatchUpDelay time.Duration

	// Maximal number of bytes used for message backlogging buffers
	// (only message payloads are counted towards MsgBufCapacity).
	// Same as Config.MsgBufCapacity, but used only for one instance of PBFT.
	// Must not be negative.
	MsgBufCapacity int

	// The maximal number of requests in a proposed request batch.
	// As soon as the number of pending requests reaches MaxBatchSize,
	// the PBFT instance may decide to immediately propose a new request batch.
	// Setting MaxBatchSize to zero signifies no limit on batch size.
	MaxBatchSize t.NumRequests

	// Per-batch view change timeout for view 0.
	// If no batch is delivered by a PBFT instance within this timeout, the node triggers a view change.
	// With each new view, the timeout doubles (without changing this value)
	ViewChangeBatchTimeout time.Duration

	// View change timeout for view 0 for the whole segment.
	// If not all batches of the associated segment are delivered by a PBFT instance within this timeout,
	// the node triggers a view change.
	// With each new view, the timeout doubles (without changing this value)
	ViewChangeSegmentTimeout time.Duration

	// Time period between resending a ViewChange message.
	// ViewChange messages need to be resent periodically to preserve liveness.
	// Otherwise, the system could get stuck if a ViewChange message is dropped by the network.
	ViewChangeResendPeriod time.Duration
}
