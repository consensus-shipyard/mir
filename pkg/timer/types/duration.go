package types

import (
	"time"

	"github.com/filecoin-project/mir/pkg/serializing"
)

// Duration represents an interval of real time.
type Duration time.Duration

// Pb converts a TimeDuration to a type used in a Protobuf message.
func (td Duration) Pb() uint64 {
	return uint64(td)
}

func (td Duration) Bytes() []byte {
	return serializing.Uint64ToBytes(uint64(td))
}
