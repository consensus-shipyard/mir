package debugger

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/logging"
	t "github.com/filecoin-project/mir/pkg/types"
)

// NewWebSocketDebugger initializes the interceptor for a given node and uses the given port for the WebSocket connection
func NewWebSocketDebugger(
	ownID t.NodeID,
	port string,
	logger logging.Logger,
) (*eventlog.Recorder, error) {
	// writerFactory creates and returns a WebSocket-based event writer
	writerFactory := func(_ string, ownID t.NodeID, l logging.Logger) (eventlog.EventWriter, error) {
		return newWSWriter(fmt.Sprintf(":%s", port), l), nil
	}

	var interceptor *eventlog.Recorder
	var err error
	interceptor, err = eventlog.NewRecorder(
		ownID,
		fmt.Sprintf("./node%s", ownID),
		logger,
		eventlog.EventWriterOpt(writerFactory),
		eventlog.SyncWriteOpt(),
	)
	if err != nil {
		return nil, err
	}
	return interceptor, err
}
