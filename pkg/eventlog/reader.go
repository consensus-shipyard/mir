package eventlog

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
)

type Reader struct {
	buffer   *bytes.Buffer
	gzReader *gzip.Reader
	source   *bufio.Reader
}

func NewReader(source io.Reader) (*Reader, error) {
	gzReader, err := gzip.NewReader(source)
	if err != nil {
		return nil, errors.WithMessage(err, "could not read source as a gzip stream")
	}

	return &Reader{
		buffer:   &bytes.Buffer{},
		gzReader: gzReader,
		source:   bufio.NewReader(gzReader),
	}, nil
}

func (r *Reader) ReadEntry() (*recordingpb.Entry, error) {
	re := &recordingpb.Entry{}
	err := readSizePrefixedProto(r.source, re, r.buffer)
	if errors.Is(err, io.EOF) {
		r.gzReader.Close()
		return re, err
	}
	if err != nil {
		return nil, errors.WithMessage(err, "error reading event")
	}
	r.buffer.Reset()

	return re, nil
}

func (r *Reader) ReadAllEvents() ([]*eventpb.Event, error) {
	allEvents := make([]*eventpb.Event, 0)

	var entry *recordingpb.Entry
	var err error
	for entry, err = r.ReadEntry(); err == nil; entry, err = r.ReadEntry() {
		allEvents = append(allEvents, entry.Events...)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return allEvents, nil
}

func readSizePrefixedProto(reader *bufio.Reader, msg proto.Message, buffer *bytes.Buffer) error {
	l, err := binary.ReadVarint(reader)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return err
		}
		return errors.WithMessage(err, "could not read size prefix")
	}

	buffer.Grow(int(l))

	if _, err := io.CopyN(buffer, reader, l); err != nil {
		return errors.WithMessage(err, "could not read message")
	}

	if err := proto.Unmarshal(buffer.Bytes(), msg); err != nil {
		return errors.WithMessage(err, "could not unmarshal message")
	}

	return nil
}
