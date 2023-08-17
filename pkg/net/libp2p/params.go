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
}

func DefaultParams() Params {
	return Params{
		ProtocolID:           "/mir/0.0.1",
		ConnectionTTL:        DefaultConnectionTTL,
		ConnectionBufferSize: 128,
		StreamWriteTimeout:   100 * time.Millisecond,
		ReconnectionPeriod:   time.Second,
	}
}
