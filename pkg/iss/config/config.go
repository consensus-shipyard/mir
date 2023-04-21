/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"fmt"
	"time"

	lsp "github.com/filecoin-project/mir/pkg/iss/leaderselectionpolicy"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
)

// The ModuleParams type defines all the ISS configuration parameters.
// Note that some fields specify delays in ticks of the logical clock.
// To obtain real time delays, these need to be multiplied by the period of the ticker provided to the Node at runtime.
type ModuleParams struct {

	// The identities of all nodes that execute the protocol in the first epoch.
	// Must not be empty.
	InitialMembership *commonpbtypes.Membership

	// Number of epochs by which to delay configuration changes.
	// If a configuration is agreed upon in epoch e, it will take effect in epoch e + 1 + configOffset.
	// Thus, in the "current" configuration, ConfigOffset subsequent configurations are already known.
	ConfigOffset int

	// The length of an ISS segment, in sequence numbers.
	// This is the number of commitLog entries each orderer needs to output in an epoch.
	// Depending on the number of leaders (and thus orderers), this will result in epoch of different lengths.
	// If set to 0, the EpochLength parameter must be non-zero and will be used to calculate the length of the segments
	// such that their lengths sum up to EpochLength.
	// Must not be negative.
	SegmentLength int

	// The length of an ISS epoch, in sequence numbers.
	// If EpochLength is non-zero, the epoch will always have a fixed
	// length, regardless of the number of leaders.
	// In each epoch, the corresponding segment lengths will be calculated to sum up to EpochLength,
	// potentially resulting in different segment length across epochs as well as within an epoch.
	// If set to zero, SegmentLength must be non-zero and will be used directly to set the length of each segment.
	// Must not be negative.
	// TODO: That EpochLength is not implemented now. SegmentLength has to be used.
	EpochLength int

	// The maximum time duration between two proposals of an orderer, where applicable.
	// For orderers that wait for an availability certificate to fill before proposing it (e.g. PBFT),
	// this parameter caps the waiting time in order to bound latency.
	// When MaxProposeDelay has elapsed since the last proposal made by an orderer,
	// the orderer proposes a new availability certificate.
	// Must not be negative.
	MaxProposeDelay time.Duration

	// Total number of buckets used by ISS.
	// In each epoch, these buckets are re-distributed evenly among the orderers.
	// Must be greater than 0.
	// TODO: Remove this parameter.
	NumBuckets int

	// Number of logical time ticks to wait until demanding retransmission of missing requests.
	// If a node receives a proposal containing requests that are not in the node's buckets,
	// it cannot accept the proposal.
	// In such a case, the node will wait for RequestNAckTimeout ticks
	// before trying to fetch those requests from other nodes.
	// Must be positive.
	RequestNAckTimeout int

	// Maximal number of bytes used for message backlogging buffers
	// (only message payloads are counted towards MsgBufCapacity).
	// On reception of a message that the node is not yet ready to process
	// (e.g., a message from a future epoch received from another node that already transitioned to that epoch),
	// the message is stored in a buffer for later processing (e.g., when this node also transitions to that epoch).
	// This total buffer capacity is evenly split among multiple buffers, one for each node,
	// so that one misbehaving node cannot exhaust the whole buffer space.
	// The most recently received messages that together do not exceed the capacity are stored.
	// If the capacity is set to 0, all messages that cannot yet be processed are dropped on reception.
	// Must not be negative.
	MsgBufCapacity int

	// Number of most recent epochs that are older than the latest stable checkpoint.
	RetainedEpochs int

	// Every CatchUpTimerPeriod, a node checks whether other nodes have fallen behind
	// and, if so, sends them the latest state.
	CatchUpTimerPeriod time.Duration

	// Time interval for repeated retransmission of checkpoint messages.
	CheckpointResendPeriod time.Duration

	// Leader selection policy to use for generating the initial state snapshot. See type for possible options.
	// This field is only used for initialization of the state snapshot that might be provided to a new instance of ISS.
	// It is the state snapshot (and not this value) that determines the protocol instance's leader selection policy.
	LeaderSelectionPolicy lsp.LeaderPolicyType

	// View change timeout for the PBFT sub-protocol, in ticks.
	// TODO: Separate this in a sub-group of the ISS params, maybe even use a field of type PBFTConfig in ModuleParams.
	PBFTDoneResendPeriod         time.Duration
	PBFTCatchUpDelay             time.Duration
	PBFTViewChangeSNTimeout      time.Duration
	PBFTViewChangeSegmentTimeout time.Duration
	PBFTViewChangeResendPeriod   time.Duration
}

// CheckParams checks whether the given configuration satisfies all necessary constraints.
func CheckParams(c *ModuleParams) error {

	// The membership must not be empty.
	if len(c.InitialMembership.Nodes) == 0 {
		return fmt.Errorf("empty membership")
	}

	// Check that ConfigOffset is at least 1
	if c.ConfigOffset < 1 {
		return fmt.Errorf("config offset must be at least 1")
	}

	// Segment length must not be negative.
	if c.SegmentLength < 0 {
		return fmt.Errorf("negative SegmentLength: %d", c.SegmentLength)
	}

	// Epoch length must not be negative.
	if c.EpochLength < 0 {
		return fmt.Errorf("negative EpochLength: %d", c.EpochLength)
	}

	// Exactly one of SegmentLength and EpochLength must be zero.
	// (The one that is zero is computed based on the non-zero one.)
	if (c.EpochLength != 0 && c.SegmentLength != 0) || (c.EpochLength == 0 && c.SegmentLength == 0) {
		return fmt.Errorf("conflicting EpochLength (%d) and SegmentLength (%d) (exactly one must be zero)",
			c.EpochLength, c.SegmentLength)
	}

	// MaxProposeDelay must not be negative.
	if c.MaxProposeDelay < 0 {
		return fmt.Errorf("negative MaxProposeDelay: %v", c.MaxProposeDelay)
	}

	// There must be at least one bucket.
	if c.NumBuckets <= 0 {
		return fmt.Errorf("non-positive number of buckets: %d", c.NumBuckets)
	}

	// RequestNackTimeout must be positive.
	if c.RequestNAckTimeout <= 0 {
		return fmt.Errorf("non-positive RequestNAckTimeout: %d", c.RequestNAckTimeout)
	}

	// MsgBufCapacity must not be negative.
	if c.MsgBufCapacity < 0 {
		return fmt.Errorf("negative MsgBufCapacity: %d", c.MsgBufCapacity)
	}

	if c.CatchUpTimerPeriod <= 0 {
		return fmt.Errorf("non-positive CatchUpTimerPeriod: %d", c.CatchUpTimerPeriod)
	}

	// If all checks passed, return nil error.
	return nil
}

// DefaultParams returns the default configuration for a given membership.
// There is no guarantee that this configuration ensures good performance, but it will pass the CheckParams test.
// DefaultParams is intended for use during testing and hello-world examples.
// A proper deployment is expected to craft a custom configuration,
// for which DefaultParams can serve as a starting point.
func DefaultParams(initialMembership *commonpbtypes.Membership) *ModuleParams {

	// Define auxiliary variables for segment length and maximal propose delay.
	// PBFT view change timeouts can then be computed relative to those.
	return (&ModuleParams{
		InitialMembership:     initialMembership,
		ConfigOffset:          2,
		SegmentLength:         4,
		NumBuckets:            len(initialMembership.Nodes),
		RequestNAckTimeout:    16,
		MsgBufCapacity:        32 * 1024 * 1024, // 32 MiB
		RetainedEpochs:        1,
		LeaderSelectionPolicy: lsp.Simple,
	}).AdjustSpeed(time.Second)
}

// AdjustSpeed sets multiple ISS parameters (e.g. view change timeouts)
// to their default values relative to maxProposeDelay.
// It can be useful to make the whole protocol run faster or slower.
// For example, for a large maxProposeDelay, the view change timeouts must be increased correspondingly,
// otherwise the view change can kick in before a node makes a proposal.
// AdjustSpeed makes these adjustments automatically.
func (mp *ModuleParams) AdjustSpeed(maxProposeDelay time.Duration) *ModuleParams {
	mp.MaxProposeDelay = maxProposeDelay
	mp.CatchUpTimerPeriod = maxProposeDelay
	mp.PBFTDoneResendPeriod = maxProposeDelay
	mp.PBFTCatchUpDelay = maxProposeDelay
	mp.CheckpointResendPeriod = maxProposeDelay
	mp.PBFTViewChangeSNTimeout = 4 * maxProposeDelay
	mp.PBFTViewChangeSegmentTimeout = 2 * time.Duration(mp.SegmentLength) * maxProposeDelay
	mp.PBFTViewChangeResendPeriod = maxProposeDelay
	// TODO: Adapt this if needed when PBFT configuration is specified separately.

	return mp
}

func StrongQuorum(n int) int {
	// assuming n > 3f:
	//   return min q: 2q > n+f
	f := MaxFaulty(n)
	return (n+f)/2 + 1
}

func WeakQuorum(n int) int {
	// assuming n > 3f:
	//   return min q: q > f
	return MaxFaulty(n) + 1
}

func MaxFaulty(n int) int {
	// assuming n > 3f:
	//   return max f
	return (n - 1) / 3
}
