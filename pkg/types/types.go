/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"encoding/binary"
	"strconv"
	"time"

	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// ================================================================================

// NodeAddress represents the address of a node.
type NodeAddress multiaddr.Multiaddr

func MembershipPb(membership map[NodeID]NodeAddress) *commonpb.Membership {
	nodeAddresses := make(map[string]string, len(membership))
	maputil.IterateSorted(membership, func(nodeID NodeID, nodeAddr NodeAddress) (cont bool) {
		nodeAddresses[nodeID.Pb()] = nodeAddr.String()
		return true
	})
	return &commonpb.Membership{Membership: nodeAddresses}
}

func Membership(membershipPb *commonpb.Membership) map[NodeID]NodeAddress {
	membership := make(map[NodeID]NodeAddress)
	for _, nID := range maputil.GetKeys(membershipPb.Membership) {
		membership[NodeID(nID)], _ = multiaddr.NewMultiaddr(membershipPb.Membership[nID])
		// TODO: Handle errors here!!!
	}
	return membership
}

// ================================================================================

// NodeID represents the ID of a node.
type NodeID string

func NewNodeIDFromInt(id int) NodeID {
	return NodeID(strconv.Itoa(id))
}

// Pb converts a NodeID to its underlying native type.
func (id NodeID) Pb() string {
	return string(id)
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

// TxID is a unique identifier of a transaction.
type TxID string

// Pb converts a TxID to its underlying native type.
func (id TxID) Pb() []byte {
	return []byte(id)
}

// Bytes serializes the object to a sequence of bytes.
func (id TxID) Bytes() []byte {
	return []byte(id)
}

// TxIDSlice converts a slice of TxIDs represented directly as their underlying native type
// to a slice of abstractly typed transaction IDs.
func TxIDSlice(ids [][]byte) []TxID {
	txIDs := make([]TxID, len(ids))
	for i, nid := range ids {
		txIDs[i] = TxID(nid)
	}
	return txIDs
}

// TxIDSlicePb converts a slice of TxIDs to a slice of the native type underlying TxID.
// This is required for serialization using Protocol Buffers.
func TxIDSlicePb(ids []TxID) [][]byte {
	pbSlice := make([][]byte, len(ids))
	for i, nid := range ids {
		pbSlice[i] = nid.Pb()
	}
	return pbSlice
}

// ================================================================================

// BatchID is a unique identifier of a batch.
type BatchID string

// Pb converts a BatchID to its underlying native type.
func (id BatchID) Pb() []byte {
	return []byte(id)
}

// Bytes serializes the object to a sequence of bytes.
func (id BatchID) Bytes() []byte {
	return []byte(id)
}

// BatchIDSlice converts a slice of BatchIDs represented directly as their underlying native type
// to a slice of abstractly typed batch IDs.
func BatchIDSlice(ids [][]byte) []BatchID {
	batchIDs := make([]BatchID, len(ids))
	for i, nid := range ids {
		batchIDs[i] = BatchID(nid)
	}
	return batchIDs
}

// BatchIDSlicePb converts a slice of BatchIDs to a slice of the native type underlying BatchID.
// This is required for serialization using Protocol Buffers.
func BatchIDSlicePb(ids []BatchID) [][]byte {
	pbSlice := make([][]byte, len(ids))
	for i, nid := range ids {
		pbSlice[i] = nid.Pb()
	}
	return pbSlice
}

// ================================================================================

// ErrorPb converts a error to its representation in protocol buffers.
func ErrorPb(err error) (ok bool, errStr string) {
	ok = err == nil
	if ok {
		errStr = ""
	} else {
		errStr = err.Error()
	}
	return
}

func ErrorFromPb(ok bool, errStr string) error {
	if ok {
		return nil
	}
	return errors.New(errStr)
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

// ================================================================================

// WALRetIndex represents the WAL (Write-Ahead Log) retention index assigned to every entry (and used for truncating).
type WALRetIndex uint64

// Pb converts a WALRetIndex to its underlying native type.
func (wri WALRetIndex) Pb() uint64 {
	return uint64(wri)
}

// ================================================================================

// TimerRetIndex represents the Timer retention index used for garbage-collecting "Repeat" invocations.
type TimerRetIndex uint64

// Pb converts a TimerRetIndex to its underlying native type.
func (tri TimerRetIndex) Pb() uint64 {
	return uint64(tri)
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

func (e EpochNr) Bytes() []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(e))
	return bytes
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

// TimeDuration represents an interval of real time
type TimeDuration time.Duration

// Pb converts a TimeDuration to a type used in a Protobuf message.
func (td TimeDuration) Pb() uint64 {
	return uint64(td)
}

// ================================================================================

// ModuleID represents an identifier of a module.
// Modules are stored under their identifiers in modules.Modules
type ModuleID string

// Pb converts a ModuleID to a type used in a Protobuf message.
func (mid ModuleID) Pb() string {
	return string(mid)
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
