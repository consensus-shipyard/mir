/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"strconv"

	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
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
