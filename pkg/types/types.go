/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"strconv"

	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

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
// Auxiliary functions
// ================================================================================
