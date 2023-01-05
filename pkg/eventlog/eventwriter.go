package eventlog

import (
	"compress/gzip"

	"github.com/filecoin-project/mir/pkg/logging"
	t "github.com/filecoin-project/mir/pkg/types"
)

type EventWriter interface {
	Write(record EventRecord) error
	Close() error
}

// DefaultNewEventWriter returns the default event writer.
// It returns the Gzip writer, making it the default event writer.
// In empirical tests comparing compression levels, best speed was only a few tenths of a percent
// worse than best compression, but your results may vary.
var DefaultNewEventWriter = func(dest string, nodeID t.NodeID, logger logging.Logger) (EventWriter, error) {
	return NewGzipWriter(dest, gzip.BestSpeed, nodeID, logger)
}
