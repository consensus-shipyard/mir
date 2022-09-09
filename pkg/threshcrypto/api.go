/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package threshcrypto

// The ThreshCrypto interface represents an implementation of the cryptographic primitives inside the MirModule module.
// It is responsible for producing and verifying cryptographic threshold signatures.
// It internally stores information about the group's public key and the node's private key share.
type ThreshCrypto interface {

	// SignShare signs the provided data and returns the resulting signature share.
	// The data to be signed is the concatenation of all the passed byte slices.
	// A signature share produced by SignShare is verifiable using VerifyShare.
	// After combining the required amount of signature shares, they can be combined into a full signature using Combine.
	// Note that the private key used to produce the signature cannot be set ("registered") through this interface.
	// Storing and using the private key is completely implementation-dependent.
	SignShare(data [][]byte) ([]byte, error)

	VerifyShare(data [][]byte, signatureShare []byte) error

	Recover(data [][]byte, signatureShares [][]byte) ([]byte, error)

	// VerifyFull verifies a full signature produced by the node with ID nodeID over data.
	// Returns nil on success (i.e., if the given signature is valid) and a non-nil error otherwise.
	VerifyFull(data [][]byte, signature []byte) error
}
