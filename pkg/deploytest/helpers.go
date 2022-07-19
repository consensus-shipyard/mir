package deploytest

import (
	"github.com/filecoin-project/mir/pkg/logging"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// NewNodeIDs returns a slice of node ids of the given size suitable for testing.
func NewNodeIDs(nNodes int) []t.NodeID {
	return sliceutil.Generate(nNodes, func(i int) t.NodeID {
		// TODO: use non-integer node IDs once all calls to strconv.Atoi are cleaned up.
		return t.NewNodeIDFromInt(i)
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
