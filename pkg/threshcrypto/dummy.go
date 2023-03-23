package threshcrypto

import (
	"bytes"
	"fmt"

	t "github.com/filecoin-project/mir/pkg/types"
)

// DummyCrypto represents a dummy MirModule module that
// always produces the same dummy byte slices specified at instantiation as the full signature.
// Signature shares always consist of the nodeID followed by a preset suffix (DummySigShareSuffix)
// Verification of these dummy signatures always succeeds.
// This is intended as a stub for testing purposes.
type DummyCrypto struct {
	// The only accepted signature share suffix
	DummySigShareSuffix []byte

	// Current node ID
	Self t.NodeID

	// The only accepted full signature
	DummySigFull []byte
}

// SignShare always returns the dummy signature DummySig, regardless of the data.
func (dc *DummyCrypto) SignShare(_ [][]byte) ([]byte, error) {
	return dc.buildSigShare(dc.Self), nil
}

// VerifyShare returns nil (i.e. success) only if signature share equals nodeID||DummySigShareSuffix.
// data is ignored.
func (dc *DummyCrypto) VerifyShare(_ [][]byte, sigShare []byte, nodeID t.NodeID) error {
	if !bytes.Equal(sigShare, dc.buildSigShare(nodeID)) {
		return fmt.Errorf("dummy signature mismatch")
	}

	return nil
}

// VerifyFull returns nil (i.e. success) only if signature equals DummySig.
// data is ignored.
func (dc *DummyCrypto) VerifyFull(_ [][]byte, signature []byte) error {
	if !bytes.Equal(signature, dc.DummySigFull) {
		return fmt.Errorf("dummy signature mismatch")
	}

	return nil
}

// Recovers full signature from signature shares if they are valid, otherwise an error is returned.
// data is ignored.
func (dc *DummyCrypto) Recover(data [][]byte, sigShares [][]byte) ([]byte, error) {
	for _, share := range sigShares {
		nodeID := share[:(len(share) - len(dc.DummySigShareSuffix))]

		if err := dc.VerifyShare(data, share, t.NodeID(nodeID)); err != nil {
			return nil, err
		}
	}

	return dc.DummySigFull, nil
}

// construct the dummy signature share for a nodeID
func (dc *DummyCrypto) buildSigShare(nodeID t.NodeID) []byte {
	share := []byte(nodeID)
	share = append(share, dc.DummySigShareSuffix...)
	return share
}
