/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package threshcrypto

// The ThreshCrypto interface represents an implementation of threshold cryptography primitives inside the MirModule module.
// It is responsible for producing and verifying cryptographic threshold signatures, which disperses the authority
// to sign among a group of N members, where T must sign their share for a full signature to be produced.
// It internally stores information about the group's public key and the node's private key share.
type ThreshCrypto interface {
	// SignShare signs the provided data and returns the resulting signature share.
	// The data to be signed is the concatenation of all the passed byte slices.
	// A signature share produced by SignShare is verifiable using VerifyShare.
	// After obtaining signature shares from T group members, the full signature can be constructed with Recover.
	// Returns the signature (and a nil error) on success, and a non-nil error otherwise.
	SignShare(data [][]byte) ([]byte, error)

	// VerifyShare verifies that a signature share is valid for the given data.
	// Returns nil on success (i.e., if the given signature share is valid) and a non-nil error otherwise.
	VerifyShare(data [][]byte, signatureShare []byte) error

	// Recover constructs a full signature from signature shares over data.
	// The given array may contain invalid signature shares: as long as at least T distinct signature
	// shares are valid, Recover will succeed, albeit less efficiently.
	// Returns the full signature (and a nil error) on success and a non-nil error otherwise.
	// Signatures returned by Recover are guaranteed to be valid.
	Recover(data [][]byte, signatureShares [][]byte) ([]byte, error)

	// VerifyFull verifies a full signature from the group over data.
	// Returns nil on success (i.e., if the given signature is valid) and a non-nil error otherwise.
	VerifyFull(data [][]byte, signature []byte) error
}
