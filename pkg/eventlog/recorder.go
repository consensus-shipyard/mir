/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// TODO: Properly comment all the code in this file.
//       Also revise the existing comments and make sure they are consistent and understandable.

package eventlog

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type EventRecord struct {
	Events *events.EventList
	Time   int64
}

// Recorder is intended to be used as an implementation of the
// mir.EventInterceptor interface.  It receives state Events,
// serializes them, compresses them, and writes them to a stream.
type Recorder struct {
	nodeID            t.NodeID
	dest              *os.File
	timeSource        func() int64
	compressionLevel  int
	retainRequestData bool
	newDests          func(EventRecord) []EventRecord
	path              string
	filter            func(event *eventpb.Event) bool

	logger    logging.Logger
	eventC    chan EventRecord
	doneC     chan struct{}
	exitC     chan struct{}
	fileCount int

	exitErr      error
	exitErrMutex sync.Mutex
}

func NewRecorder(
	nodeID t.NodeID,
	path string,
	logger logging.Logger,
	opts ...RecorderOpt,
) (*Recorder, error) {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	startTime := time.Now()

	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, fmt.Errorf("error creating interceptor directory: %w", err)
	}

	dest, err := os.Create(filepath.Join(path, "eventlog0.gz"))
	if err != nil {
		return nil, fmt.Errorf("error creating event log file: %w", err)
	}

	i := &Recorder{
		dest:   dest,
		nodeID: nodeID,
		timeSource: func() int64 {
			return time.Since(startTime).Milliseconds()
		},
		compressionLevel: DefaultCompressionLevel,
		eventC:           make(chan EventRecord, DefaultBufferSize),
		doneC:            make(chan struct{}),
		exitC:            make(chan struct{}),
		fileCount:        1,
		newDests:         OneFileLogger(),
		path:             path,
		filter: func(event *eventpb.Event) bool {
			// Record all events by default.
			return true
		},
		logger: logger,
	}

	for _, opt := range opts {
		switch v := opt.(type) {
		case timeSourceOpt:
			i.timeSource = v
		case retainRequestDataOpt:
			i.retainRequestData = true
		case compressionLevelOpt:
			i.compressionLevel = int(v)
		case bufferSizeOpt:
			i.eventC = make(chan EventRecord, v)
		case fileSplitterOpt:
			i.newDests = v
		case eventFilterOpt:
			oldFilter := i.filter
			i.filter = func(e *eventpb.Event) bool {
				// Apply the given filter on top of the existing one.
				return oldFilter(e) && v(e)
			}
		}
	}

	go func() {
		err := i.run()
		if err != nil {
			logger.Log(logging.LevelError, "Interceptor returned with error.", "err", err)
		} else {
			logger.Log(logging.LevelDebug, "Interceptor returned successfully.")
		}
	}()

	return i, nil
}

// Intercept takes an event and enqueues it into the event buffer.
// If there is no room in the buffer, it blocks.  If draining the buffer
// to the output stream has completed (successfully or otherwise), Intercept
// returns an error.
func (i *Recorder) Intercept(events *events.EventList) error {
	select {
	case i.eventC <- EventRecord{
		Events: events,
		Time:   i.timeSource(),
	}:
		return nil
	case <-i.exitC:
		i.exitErrMutex.Lock()
		defer i.exitErrMutex.Unlock()
		return i.exitErr
	}
}

// Stop must be invoked to release the resources associated with this
// Interceptor, and should only be invoked after the mir node has completely
// exited.  The returned error
func (i *Recorder) Stop() error {

	// Send stop signal to writer thread and wait for it to stop.
	close(i.doneC)
	<-i.exitC

	// Close destination file.
	err := i.dest.Close()
	if err != nil {
		i.logger.Log(logging.LevelError, "Failed to close event recorder output stream.", "error", err)
	}

	// Return final error if it is different from the expected errStopped.
	i.exitErrMutex.Lock()
	defer i.exitErrMutex.Unlock()
	if errors.Is(i.exitErr, errStopped) {
		return nil
	}
	return i.exitErr
}

var errStopped = fmt.Errorf("interceptor stopped at caller request")

func (i *Recorder) run() (exitErr error) {
	cnt := 0 // Counts total number of Events written.

	defer func() {
		i.exitErrMutex.Lock()
		i.exitErr = exitErr
		i.exitErrMutex.Unlock()
		close(i.exitC)
		i.logger.Log(logging.LevelInfo, "Intercepted Events written to event log.", "numEvents", cnt)
	}()

	write := func(dest io.Writer, record EventRecord) error {

		gzWriter, err := gzip.NewWriterLevel(dest, i.compressionLevel)
		if err != nil {
			return err
		}
		defer func() {
			if err := gzWriter.Close(); err != nil {
				i.logger.Log(logging.LevelError, "Error closing gzWriter.", "err", err)
			}
		}()

		return writeRecordedEvent(gzWriter, &recordingpb.Entry{
			NodeId: i.nodeID.Pb(),
			Time:   record.Time,
			Events: filter(record.Events, i.filter).Slice(),
		})
	}

	writeInFiles := func(record EventRecord) error {
		eventByDests := i.newDests(record)
		count := 0
		for _, rec := range eventByDests {
			if count > 0 {
				// newDest required
				dest, err := os.Create(filepath.Join(i.path, "eventlog"+strconv.Itoa(i.fileCount)+".gz"))
				if err != nil {
					return err
				}
				err = i.dest.Close()
				if err != nil {
					return err
				}
				i.dest = dest
				i.fileCount++
			}

			if rec.Events.Len() != 0 {
				if err := write(i.dest, rec); err != nil {
					return err
				}
			}
			count++
		}
		return nil
	}

	for {
		select {
		case <-i.doneC:
			for {
				select {
				case record := <-i.eventC:

					if err := writeInFiles(record); err != nil {
						return errors.WithMessage(err, "error serializing to stream")
					}
				default:
					return errStopped
				}
			}
		case record := <-i.eventC:
			if err := writeInFiles(record); err != nil {
				return errors.WithMessage(err, "error serializing to stream")
			}
			cnt++
		}
	}
}

func writeRecordedEvent(writer io.Writer, entry *recordingpb.Entry) error {
	return writeSizePrefixedProto(writer, entry)
}

func writeSizePrefixedProto(dest io.Writer, msg proto.Message) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return errors.WithMessage(err, "could not marshal")
	}

	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(lenBuf, int64(len(msgBytes)))
	if _, err = dest.Write(lenBuf[:n]); err != nil {
		return errors.WithMessage(err, "could not write length prefix")
	}

	if _, err = dest.Write(msgBytes); err != nil {
		return errors.WithMessage(err, "could not write message")
	}

	return nil
}

func filter(evts *events.EventList, predicate func(event *eventpb.Event) bool) *events.EventList {
	filtered := &events.EventList{}

	iter := evts.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		if predicate(event) {
			filtered.PushBack(event)
		}
	}

	return filtered
}
