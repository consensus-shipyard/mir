package eventlog

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type RecorderOpt interface{}

type timeSourceOpt func() int64

// TimeSourceOpt can be used to override the default Time source
// for an interceptor.  This can be useful for changing the
// granularity of the timestamps, or picking some externally
// supplied sync point when trying to synchronize logs.
// The default Time source will timestamp with the Time, in
// milliseconds since the interceptor was created.
func TimeSourceOpt(source func() int64) RecorderOpt {
	return timeSourceOpt(source)
}

// DefaultBufferSize is the number of unwritten state Events which
// may be held in queue before blocking.
const DefaultBufferSize = 5000

type bufferSizeOpt int

// BufferSizeOpt overrides the default buffer size of the
// interceptor buffer.  Once the buffer overflows, the state
// machine will be blocked from receiving new state Events
// until the buffer has room.
func BufferSizeOpt(size int) RecorderOpt {
	return bufferSizeOpt(size)
}

// TODO: Write documenting comments for the options below.

type fileSplitterOpt func(EventRecord) []EventRecord

func FileSplitterOpt(splitter func(EventRecord) []EventRecord) RecorderOpt {
	return fileSplitterOpt(splitter)
}

type eventFilterOpt func(*eventpb.Event) bool

func EventFilterOpt(filter func(event *eventpb.Event) bool) RecorderOpt {
	return eventFilterOpt(filter)
}

type eventWriterOpt func(dest string, nodeID t.NodeID, logger logging.Logger) (EventWriter, error)

func EventWriterOpt(
	factory func(dest string, nodeID t.NodeID, logger logging.Logger) (EventWriter, error),
) RecorderOpt {
	return eventWriterOpt(factory)
}
