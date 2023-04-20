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

// SeqNrSlice converts a slice of SeqNrs represented directly as their underlying native type
// to a slice of abstractly typed sequence numbers.
func SeqNrSlice(sns []uint64) []SeqNr {
	seqNrs := make([]SeqNr, len(sns))
	for i, nid := range sns {
		seqNrs[i] = SeqNr(nid)
	}
	return seqNrs
}

// SeqNrSlicePb converts a slice of SeqNrs to a slice of the native type underlying SeqNr.
// This is required for serialization using Protocol Buffers.
func SeqNrSlicePb(sns []SeqNr) []uint64 {
	pbSlice := make([]uint64, len(sns))
	for i, sn := range sns {
		pbSlice[i] = sn.Pb()
	}
	return pbSlice
}

// ================================================================================

// ReqNo represents a request number a client assigns to its requests.
type ReqNo uint64

// Pb converts a ReqNo to its underlying native type.
func (rn ReqNo) Pb() uint64 {
	return uint64(rn)
}

// Bytes converts a ReqNo to a slice of bytes (useful for serialization).
func (rn ReqNo) Bytes() []byte {
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
type TxID = []byte

// TxIDSlice converts a slice of TxIDs represented directly as their underlying native type
// to a slice of abstractly typed transaction IDs.
func TxIDSlice(ids [][]byte) []TxID {
	txIDs := make([]TxID, len(ids))
	copy(txIDs, ids)
	return txIDs
}

// TxIDSlicePb converts a slice of TxIDs to a slice of the native type underlying TxID.
// This is required for serialization using Protocol Buffers.
func TxIDSlicePb(ids []TxID) [][]byte {
	pbSlice := make([][]byte, len(ids))
	copy(pbSlice, ids)
	return pbSlice
}

// ================================================================================
