package eventlog

import (
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"

	es "github.com/go-errors/errors"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

type gzipWriter struct {
	dest             *os.File
	compressionLevel int
	nodeID           t.NodeID
	logger           logging.Logger
}

func NewGzipWriter(filename string, compressionLevel int, nodeID t.NodeID, logger logging.Logger) (EventWriter, error) {
	dest, err := os.Create(filename + ".gz")
	if err != nil {
		return nil, es.Errorf("error creating event log file: %w", err)
	}
	return &gzipWriter{
		dest:             dest,
		compressionLevel: compressionLevel,
		nodeID:           nodeID,
		logger:           logger,
	}, nil
}

func (w *gzipWriter) Write(record EventRecord) error {
	gzWriter, err := gzip.NewWriterLevel(w.dest, w.compressionLevel)
	if err != nil {
		return err
	}
	defer func() {
		if err := gzWriter.Close(); err != nil {
			w.logger.Log(logging.LevelError, "Error closing gzWriter.", "err", err)
		}
	}()

	return writeRecordedEvent(gzWriter, &recordingpb.Entry{
		NodeId: w.nodeID.Pb(),
		Time:   record.Time,
		Events: sliceutil.Transform(record.Events.Slice(), func(_ int, ev *eventpbtypes.Event) *eventpb.Event { return ev.Pb() }),
	})
}

func (w *gzipWriter) Flush() error {
	return w.dest.Sync()
}

func (w *gzipWriter) Close() error {
	return w.dest.Close()
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
