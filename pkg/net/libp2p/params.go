package libp2p

import (
	"math"
	"time"

	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	DefaultConnectionTTL = time.Duration(math.MaxInt64 - iota)
)

type Params struct {
	ProtocolID           protocol.ID
	ConnectionTTL        time.Duration `json:",string"`
	ConnectionBufferSize int           `json:",string"`
	StreamWriteTimeout   time.Duration `json:",string"`
	ReconnectionPeriod   time.Duration `json:",string"`
	MinComplainPeriod    time.Duration `json:",string"`

	// For some strange reason, a call to Write() on a libp2p stream tends to pretend to time out (without the timeout
	// actually elapsing) if we provide a bigger amount of data (hundreds of kB) at once. Splitting it into smaller
	// chunks and calling Write() successively with these chunks seems to resolve the issue.
	// Experienced on a multi-continental deployment on AWS EC2, with libp2p v0.26.2 and yamux v4.0.0.
	MaxDataPerWrite int `json:",string"`
}

func DefaultParams() Params {
	return Params{
		ProtocolID:           "/mir/0.0.1",
		ConnectionTTL:        DefaultConnectionTTL,
		ConnectionBufferSize: 512,
		StreamWriteTimeout:   100 * time.Millisecond,
		ReconnectionPeriod:   3 * time.Second,
		MaxDataPerWrite:      100 * 1024, // 100 kiB
		MinComplainPeriod:    time.Second,
	}
}
