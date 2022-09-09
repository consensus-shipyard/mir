/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package threshcrypto

import (
	"bytes"
	"fmt"
)

// DummyCrypto represents a dummy MirModule module that
// always produces the same dummy byte slices specified at instantiation as the signature share/full signature.
// Verification of these dummy signatures always succeeds.
// This is intended as a stub for testing purposes.
type DummyCrypto struct {
	// The only accepted signature share
	DummySigShare []byte

	// The only accepted full signature
	DummySigFull []byte
}

// Sign always returns the dummy signature DummySig, regardless of the data.
func (dc *DummyCrypto) SignShare(data [][]byte) ([]byte, error) {
	return dc.DummySigShare, nil
}

// Verify returns nil (i.e. success) only if signature share equals DummySigShare.
// data is ignored.
func (dc *DummyCrypto) VerifyShare(data [][]byte, sigShare []byte) error {
	if !bytes.Equal(sigShare, dc.DummySigShare) {
		return fmt.Errorf("dummy signature mismatch")
	}

	return nil
}

// Verify returns nil (i.e. success) only if signature equals DummySig.
// data is ignored.
func (dc *DummyCrypto) VerifyFull(data [][]byte, signature []byte) error {
	if !bytes.Equal(signature, dc.DummySigFull) {
		return fmt.Errorf("dummy signature mismatch")
	}

	return nil
}

// Verifies signature shares and produces DummySigFull if they are valid, otherwise an error is returned.
// data is ignored.
func (dc *DummyCrypto) Recover(data [][]byte, sigShares [][]byte) ([]byte, error) {
	for _, share := range sigShares {
		if err := dc.VerifyShare(data, share); err != nil {
			return nil, err
		}
	}

	return dc.DummySigFull, nil
}
