/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// TODO: Properly comment all the code in this file.
//       Also revise the existing comments and make sure they are consistent and understandable.

package eventlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/pkg/errors"
)

type EventRecord struct {
	Events *events.EventList
	Time   int64
}

func (record *EventRecord) Filter(predicate func(event *eventpb.Event) bool) EventRecord {
	filtered := &events.EventList{}

	iter := record.Events.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		if predicate(event) {
			filtered.PushBack(event)
		}
	}

	return EventRecord{
		Events: filtered,
		Time:   record.Time,
	}
}

// Recorder is intended to be used as an implementation of the
// mir.EventInterceptor interface.  It receives state Events,
// serializes them, compresses them, and writes them to a stream.
type Recorder struct {
	nodeID            t.NodeID
	dest              EventWriter
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

	i := &Recorder{
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

	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, fmt.Errorf("error creating interceptor directory: %w", err)
	}

	writer, err := NewGzipWriter(filepath.Join(path, "eventlog0.gz"), i.compressionLevel, nodeID, logger)
	if err != nil {
		return nil, fmt.Errorf("error initializing event writer: %w", err)
	}
	i.dest = writer

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
		i.logger.Log(logging.LevelError, "Failed to close event writer.", "error", err)
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

	writeInFiles := func(record EventRecord) error {
		eventByDests := i.newDests(record)
		count := 0
		for _, rec := range eventByDests {
			if count > 0 {
				// newDest required
				dest, err := NewGzipWriter(
					filepath.Join(i.path, "eventlog"+strconv.Itoa(i.fileCount)+".gz"),
					i.compressionLevel,
					i.nodeID,
					i.logger,
				)
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
				if err := i.dest.Write(rec.Filter(i.filter)); err != nil {
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
