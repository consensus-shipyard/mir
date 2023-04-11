package types

import "encoding/binary"

// ViewNr represents the view number in the PBFT protocol (used as a sub-protocol of ISS)
type ViewNr uint64

// Pb converts a ViewNr to its underlying native type
func (v ViewNr) Pb() uint64 {
	return uint64(v)
}

// Bytes converts a PBFTViewNr to a slice of bytes (useful for serialization).
func (v ViewNr) Bytes() []byte {
	return uint64ToBytes(uint64(v))
}

// Encode view number.
func uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, n)
	return buf
}
