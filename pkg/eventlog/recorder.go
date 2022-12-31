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

// Recorder is intended to be used as an imlementation of the
// mir.EventInterceptor interface.  It receives state Events,
// serializes them, compresses them, and writes them to a stream.
type Recorder struct {
	nodeID            t.NodeID
	dest              *os.File
	timeSource        func() int64
	compressionLevel  int
	retainRequestData bool
	eventC            chan EventRecord
	doneC             chan struct{}
	exitC             chan struct{}
	fileCount         int
	newDests          func(EventRecord) []EventRecord
	path              string

	exitErr      error
	exitErrMutex sync.Mutex
}

func NewRecorder(
	nodeID t.NodeID,
	path string,
	logger logging.Logger,
	newDests func(EventRecord) []EventRecord,
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
		newDests:         newDests,
		path:             path,
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
	close(i.doneC)
	<-i.exitC
	i.exitErrMutex.Lock()
	defer i.exitErrMutex.Unlock()
	if errors.Is(i.exitErr, errStopped) {
		return nil
	}
	err := i.dest.Close()
	if err != nil {
		return err
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
		fmt.Printf("Intercepted Events written to event log: %d\n", cnt)
	}()

	write := func(dest io.Writer, eventTime EventRecord) error {

		gzWriter, err := gzip.NewWriterLevel(dest, i.compressionLevel)
		if err != nil {
			return err
		}
		defer func() {
			if err := gzWriter.Close(); err != nil {
				fmt.Printf("Error closing gzWriter: %v\n", err)
			}
		}()

		return writeRecordedEvent(gzWriter, &recordingpb.Entry{
			NodeId: i.nodeID.Pb(),
			Time:   eventTime.Time,
			Events: eventTime.Events.Slice(),
		})
	}

	writeInFiles := func(event EventRecord) error {
		eventByDests := i.newDests(event)
		count := 0
		for _, eventTime := range eventByDests {
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

			if eventTime.Events.Len() != 0 {
				if err := write(i.dest, event); err != nil {
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
				case event := <-i.eventC:

					if err := writeInFiles(event); err != nil {
						return errors.WithMessage(err, "error serializing to stream")
					}
				default:
					return errStopped
				}
			}
		case event := <-i.eventC:
			if err := writeInFiles(event); err != nil {
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
