/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"encoding/binary"
	"strconv"

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

func MembershipSlice(membershipsPb []*commonpb.Membership) []map[NodeID]NodeAddress {
	memberships := make([]map[NodeID]NodeAddress, len(membershipsPb))
	for i, membershipPb := range membershipsPb {
		memberships[i] = Membership(membershipPb)
	}
	return memberships
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

// Bytes returns the byte slice representation of the node ID.
func (id NodeID) Bytes() []byte {
	return []byte(id)
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
	return Uint64ToBytes(uint64(sn))
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
	return Uint64ToBytes(uint64(rn))
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
	return Uint64ToBytes(uint64(ri))
}

// ================================================================================

// EpochNr represents the number of an epoch.
type EpochNr uint64

// Pb converts an EpochNr number to its underlying native type.
func (e EpochNr) Pb() uint64 {
	return uint64(e)
}

func (e EpochNr) Bytes() []byte {
	return Uint64ToBytes(uint64(e))
}

// ================================================================================
// Auxiliary functions
// ================================================================================

// Encode view number.
func Uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, n)
	return buf
}

func Uint64FromBytes(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}
