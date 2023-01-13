package libp2p

import (
	"math"
	"time"

	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	DefaultPermanentAddrTTL = time.Duration(math.MaxInt64 - iota)
)

type Params struct {
	ProtocolID             protocol.ID
	MaxConnectingTimeout   time.Duration
	MaxRetryTimeout        time.Duration
	MaxRetries             int
	NoLoggingErrorAttempts int
	PermanentAddrTTL       time.Duration
	ConnWaitPollInterval   time.Duration
}

func DefaultParams() Params {
	return Params{

		ProtocolID:             "/mir/0.0.1",
		MaxConnectingTimeout:   10 * time.Second,
		MaxRetryTimeout:        1 * time.Second,
		MaxRetries:             10,
		NoLoggingErrorAttempts: 2,
		PermanentAddrTTL:       DefaultPermanentAddrTTL,
		ConnWaitPollInterval:   100 * time.Millisecond,
	}
}
