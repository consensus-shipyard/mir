/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

Refactored: 1
*/

package mir

import (
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/logging"
)

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

	// Stats configures event processing statistics generation.
	// Enable by setting Stats.Period to a positive value.
	Stats StatsConfig
}

// DefaultNodeConfig returns the default node configuration.
// It can be used as a base for creating more specific configurations when instantiating a Node.
func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		Logger:               logging.ConsoleInfoLogger,
		MaxEventBatchSize:    512,
		PauseInputThreshold:  8192,
		ResumeInputThreshold: 6144,
		Stats: StatsConfig{
			Logger:   logging.ConsoleDebugLogger,
			LogLevel: logging.LevelDebug,
			Period:   0, // Stats logging disabled by default.
		},
	}
}

func (c *NodeConfig) Validate() error {
	if c.MaxEventBatchSize <= 0 {
		return errors.Errorf("MaxEventBatchSize must be greater than 0, got %d", c.MaxEventBatchSize)
	}

	if c.PauseInputThreshold <= 0 {
		return errors.Errorf("PauseInputThreshold must be greater than 0, got %d", c.PauseInputThreshold)
	}

	if c.ResumeInputThreshold < 0 {
		return errors.Errorf("ResumeInputThreshold must be greater than or equal to 0, got %d", c.ResumeInputThreshold)
	}

	if c.PauseInputThreshold < c.ResumeInputThreshold {
		return errors.Errorf("PauseInputThreshold (%d) must be greater than or equal to ResumeInputThreshold (%d)",
			c.PauseInputThreshold, c.ResumeInputThreshold)
	}

	return nil
}

func (c *NodeConfig) WithLogger(logger logging.Logger) *NodeConfig {
	newConfig := *c
	newConfig.Logger = logger
	return &newConfig
}
