/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import t "github.com/filecoin-project/mir/pkg/types"

// The Crypto interface represents an implementation of the cryptographic primitives inside the MirModule module.
// It is responsible for producing and verifying cryptographic signatures.
// It internally stores information about which clients and nodes are associated with which public keys.
// This information can be updated by the Node through appropriate events.
// Both clients and nodes are identified only by their IDs.
type Crypto interface {

	// Sign signs the provided data and returns the resulting signature.
	// The data to be signed is the concatenation of all the passed byte slices.
	// A signature produced by Sign is verifiable using Verify.
	// Note that the private key used to produce the signature cannot be set ("registered") through this interface.
	// Storing and using the private key is completely implementation-dependent.
	Sign(data [][]byte) ([]byte, error)

	// Verify verifies a signature produced by the node with ID nodeID over data.
	// Returns nil on success (i.e., if the given signature is valid) and a non-nil error otherwise.
	Verify(data [][]byte, signature []byte, nodeID t.NodeID) error
}
