package types

import (
	"strconv"

	"github.com/filecoin-project/mir/pkg/serializing"
)

// ================================================================================

// ClientID represents the ID of a client.
type ClientID string

func NewClientIDFromInt(id int) ClientID {
	return ClientID(strconv.Itoa(id))
}

// Pb converts a ClientID to its underlying native type.
func (cid ClientID) Pb() string {
	return string(cid)
}

func (cid ClientID) Bytes() []byte {
	return []byte(cid)
}

// ================================================================================

// SeqNr represents the sequence number of a batch as assigned by the ordering protocol.
type SeqNr uint64

// Pb converts a SeqNr to its underlying native type.
func (sn SeqNr) Pb() uint64 {
	return uint64(sn)
}

// Bytes converts a SeqNr to a slice of bytes (useful for serialization).
func (sn SeqNr) Bytes() []byte {
	return serializing.Uint64ToBytes(uint64(sn))
}

// ================================================================================

// TxNo represents a request number a client assigns to its requests.
type TxNo uint64

// Pb converts a TxNo to its underlying native type.
func (rn TxNo) Pb() uint64 {
	return uint64(rn)
}

// Bytes converts a TxNo to a slice of bytes (useful for serialization).
func (rn TxNo) Bytes() []byte {
	return serializing.Uint64ToBytes(uint64(rn))
}

// ================================================================================

// RetentionIndex represents an abstract notion of system progress used in garbage collection.
// The basic idea is to associate various parts of the system (parts of the state, even whole modules)
// that are subject to eventual garbage collection with a retention index.
// As the system progresses, the retention index monotonically increases
// and parts of the system associated with a lower retention index can be garbage-collected.
type RetentionIndex uint64

// Pb converts a RetentionIndex to its underlying native type.
func (ri RetentionIndex) Pb() uint64 {
	return uint64(ri)
}

// Bytes converts a RetentionIndex to a slice of bytes (useful for serialization).
func (ri RetentionIndex) Bytes() []byte {
	return serializing.Uint64ToBytes(uint64(ri))
}

// ================================================================================

// EpochNr represents the number of an epoch.
type EpochNr uint64

// Pb converts an EpochNr number to its underlying native type.
func (e EpochNr) Pb() uint64 {
	return uint64(e)
}

func (e EpochNr) Bytes() []byte {
	return serializing.Uint64ToBytes(uint64(e))
}

// ================================================================================

// TxID is a unique identifier of a transaction.
type TxID = string

// ================================================================================
