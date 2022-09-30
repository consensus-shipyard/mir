/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

Refactored: 1
*/

package mir

import "github.com/filecoin-project/mir/pkg/logging"

// The NodeConfig struct represents configuration parameters of the node
// that are independent of the protocol the Node is executing.
// NodeConfig only contains protocol-independent parameters. Protocol-specific parameters
// should be specified when instantiating the protocol implementation as one of the Node's modules.
type NodeConfig struct {
	// Logger provides the logging functions.
	Logger logging.Logger

	// MaxEventBatchSize is the maximum number of events that can be dispatched to a module in a single batch.
	MaxEventBatchSize int

	// PauseInputThreshold is the number of events in the node's internal event buffer that triggers the disabling
	// of further external input (i.e. events emitted by active modules). Events emitted by passive modules are
	// not affected. The processing of external events is resumed when the number of events in the buffer drops
	// below the ResumeInputThreshold.
	PauseInputThreshold int

	// ResumeInputThreshold is the number of events in the node's internal event buffer that triggers the enabling
	// of external input (i.e. events emitted by active modules). Events emitted by passive modules are not affected.
	// When the external input is disabled and the number of events in the buffer drops below ResumeInputThreshold,
	// external input events can be added to the event buffer again.
	ResumeInputThreshold int
}

// DefaultNodeConfig returns the default node configuration.
// It can be used as a base for creating more specific configurations when instantiating a Node.
func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		Logger:               logging.ConsoleInfoLogger,
		MaxEventBatchSize:    512,
		PauseInputThreshold:  8192,
		ResumeInputThreshold: 6144,
	}
}

func (nc *NodeConfig) WithLogger(logger logging.Logger) *NodeConfig {
	newConfig := *nc
	newConfig.Logger = logger
	return &newConfig
}
