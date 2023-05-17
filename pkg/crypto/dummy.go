/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"bytes"

	es "github.com/go-errors/errors"

	t "github.com/filecoin-project/mir/pkg/types"
)

// TODO: Write comments.

// DummyCrypto represents a dummy MirModule module that
// always produces the same dummy byte slice specified at instantiation as signature.
// Verification of this dummy signature always succeeds.
// This is intended as a stub for testing purposes.
type DummyCrypto struct {

	// The only accepted signature
	DummySig []byte
}

// Sign always returns the dummy signature DummySig, regardless of the data.
func (dc *DummyCrypto) Sign(_ [][]byte) ([]byte, error) {
	return dc.DummySig, nil
}

// RegisterNodeKey does nothing, as no public keys are used.
func (dc *DummyCrypto) RegisterNodeKey(_ []byte, _ t.NodeID) error {
	return nil
}

// DeleteNodeKey does nothing, as no public keys are used.
func (dc *DummyCrypto) DeleteNodeKey(_ t.NodeID) {
}

// Verify returns nil (i.e. success) only if signature equals DummySig.
// Both data and nodeID are ignored.
func (dc *DummyCrypto) Verify(_ [][]byte, signature []byte, _ t.NodeID) error {
	if !bytes.Equal(signature, dc.DummySig) {
		return es.Errorf("dummy signature mismatch")
	}

	return nil
}
