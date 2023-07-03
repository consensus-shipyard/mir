package deploytest

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// NewNodeIDsWeights returns a map of node ids of the given size suitable for testing with the nodeID as key and their weight as value. The weight is calculated
// with a function parameter that is iteratively called in order from the first nodeID 0 to the last nodeID nNodes-1
func NewNodeIDsWeights(nNodes int, weightFunction func(t.NodeID) types.VoteWeight) map[t.NodeID]types.VoteWeight {
	nodeIDs := make(map[t.NodeID]types.VoteWeight, nNodes)
	for i := 0; i < nNodes; i++ {
		nodeID := t.NewNodeIDFromInt(i)
		nodeIDs[nodeID] = weightFunction(nodeID)
	}
	return nodeIDs
}

func NewNodeIDsDefaultWeights(nNodes int) map[t.NodeID]types.VoteWeight {
	return NewNodeIDsWeights(nNodes, func(t.NodeID) types.VoteWeight {
		return 1
	})
}

// NewLogger returns a new logger suitable for tests.
// If parentLogger is not nil, it returns a thread-safe wrapper around parentLogger.
// Otherwise, it returns a thread-safe wrapper around logging.ConsoleDebugLogger.
func NewLogger(parentLogger logging.Logger) logging.Logger {
	if parentLogger != nil {
		return logging.Synchronize(parentLogger)
	}
	return logging.Synchronize(logging.ConsoleDebugLogger)
}
