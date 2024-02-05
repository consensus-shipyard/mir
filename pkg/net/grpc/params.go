package grpc

import (
	"math"
	"time"
)

const (
	DefaultConnectionTTL = time.Duration(math.MaxInt64 - iota)
)

type Params struct {
	ConnectionBufferSize int           `json:",string"`
	ReconnectionPeriod   time.Duration `json:",string"`
	MinComplainPeriod    time.Duration `json:",string"`
}

func DefaultParams() Params {
	return Params{
		ConnectionBufferSize: 512,
		ReconnectionPeriod:   3 * time.Second,
		MinComplainPeriod:    time.Second,
	}
}
