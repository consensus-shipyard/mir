package stdtypes

import (
	"strconv"
)

// ================================================================================

// NodeID represents the ID of a node.
type NodeID string

func NewNodeIDFromInt(id int) NodeID {
	return NodeID(strconv.Itoa(id))
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
