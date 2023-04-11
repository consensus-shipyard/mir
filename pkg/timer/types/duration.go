package types

import (
	"encoding/binary"
	"time"
)

// Duration represents an interval of real time.
type Duration time.Duration

// Pb converts a TimeDuration to a type used in a Protobuf message.
func (td Duration) Pb() uint64 {
	return uint64(td)
}

func (td Duration) Bytes() []byte {
	return uint64ToBytes(uint64(td))
}

func uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, n)
	return buf
}
