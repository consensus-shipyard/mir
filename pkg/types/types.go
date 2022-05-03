/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"encoding/binary"
	"strconv"
)

// ================================================================================

// NodeID represents the numeric ID of a node.
type NodeID string

func NewNodeIDFromInt(id int) NodeID {
	return NodeID(strconv.Itoa(id))
}

// Pb converts a NodeID to its underlying native type.
func (nid NodeID) Pb() string {
	return string(nid)
}

// NodeIDSlice converts a slice of NodeIDs represented directly as their underlying native type
// to a slice of abstractly typed node IDs.
func NodeIDSlice(nids []string) []NodeID {
	nodeIDs := make([]NodeID, len(nids))
	for i, nid := range nids {
		nodeIDs[i] = NodeID(nid)
	}
	return nodeIDs
}

// NodeIDSlicePb converts a slice of NodeIDs to a slice of the native type underlying NodeID.
// This is required for serialization using Protocol Buffers.
func NodeIDSlicePb(nids []NodeID) []string {
	pbSlice := make([]string, len(nids))
	for i, nid := range nids {
		pbSlice[i] = nid.Pb()
	}
	return pbSlice
}

// ================================================================================

// ClientID represents the numeric ID of a client.
type ClientID string

func NewClientIDFromInt(id int) ClientID {
	return ClientID(strconv.Itoa(id))
}

// Pb converts a ClientID to its underlying native type.
func (cid ClientID) Pb() string {
	return string(cid)
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
	return uint64ToBytes(uint64(sn))
}

// SeqNrSlice converts a slice of SeqNrs represented directly as their underlying native type
// to a slice of abstractly typed sequence nubmers.
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

// ================================================================================

// WALRetIndex represents the WAL (Write-Ahead Log) retention index assigned to every entry (and used for truncating).
type WALRetIndex uint64

// Pb converts a WALRetIndex to its underlying native type.
func (wri WALRetIndex) Pb() uint64 {
	return uint64(wri)
}

// ================================================================================

// SBInstanceNr identifies the instance of Sequenced Broadcast (SB) within an epoch.
type SBInstanceNr uint64

// Pb converts a SBInstanceNr to its underlying native type.
func (i SBInstanceNr) Pb() uint64 {
	return uint64(i)
}

// ================================================================================

// EpochNr represents the number of an epoch.
type EpochNr uint64

// Pb converts an EpochNr number to its underlying native type.
func (e EpochNr) Pb() uint64 {
	return uint64(e)
}

// ================================================================================

// NumRequests represents the number of requests (e.g. pending in some buffer)
type NumRequests uint64

// Pb converts an EpochNr number to its underlying native type.
func (nr NumRequests) Pb() uint64 {
	return uint64(nr)
}

// ================================================================================

// PBFTViewNr represents the view number in the PBFT protocol (used as a sub-protocol of ISS)
type PBFTViewNr uint64

// Pb converts a PBFTViewNr to its underlying native type
func (v PBFTViewNr) Pb() uint64 {
	return uint64(v)
}

// Bytes converts a PBFTViewNr to a slice of bytes (useful for serialization).
func (v PBFTViewNr) Bytes() []byte {
	return uint64ToBytes(uint64(v))
}

// ================================================================================
// Auxiliary functions
// ================================================================================

// Encode view number.
func uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, n)
	return buf
}
