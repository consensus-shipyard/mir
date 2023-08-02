package types

import (
	"math/big"
	"strconv"

	"github.com/filecoin-project/mir/pkg/serializing"
)

// ================================================================================

// VoteWeight represents the weight of a node's vote when gathering quorums.
// The underlying type is a string containing a decimal integer representation of the weight.
// This is required to support large integers that would not fit in a native data type like uint64.
// For example, this can occur if Trantor is used as a PoS system with cryptocurrency units as weights.
// We do not store the weights directly as big.Int, since that would make it harder to use them in protocol buffers.
// Instead, when performing mathematical operations on weights, we convert them to the big.Int type.
type VoteWeight string

func (vw VoteWeight) Bytes() []byte {
	return []byte(vw)
}

func (vw VoteWeight) Pb() string {
	return string(vw)
}

func (vw VoteWeight) String() string {
	return string(vw)
}

func (vw VoteWeight) IsValid() bool {
	if _, ok := new(big.Int).SetString(vw.String(), 10); !ok {
		return false
	}
	return true
}

// BigInt converts a VoteWeight (normally represented as a string) to a big.Int.
// BigInt panics if the underlying string is not a valid decimal representation of an integer.
// Thus, BigInt must not be called on received input without having validated VoteWeight by calling the IsValid method.
func (vw VoteWeight) BigInt() *big.Int {
	var bi big.Int
	if _, ok := bi.SetString(vw.String(), 10); !ok {
		panic("invalid vote weight representation")
	}
	return &bi
}

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
