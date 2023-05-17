/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"io"
	insecureRNG "math/rand"

	es "github.com/go-errors/errors"

	t "github.com/filecoin-project/mir/pkg/types"
)

var (
	// DefaultPseudoSeed is an arbitrary number that the nodes can use as a seed when instantiating its MirModule module.
	// This is not secure, but helps during testing, as it obviates the exchange of public keys among nodes.
	DefaultPseudoSeed int64 = 12345
)

// InsecureCryptoForTestingOnly returns a CryptoImpl module to be used by a node, generating new keys in a pseudo-random manner.
// It is initialized and populated deterministically, based on a given configuration and a random seed.
// InsecureCryptoForTestingOnly is not secure and intended for testing purposes only.
// It also assumes a static membership known to all nodes,
// InsecureCryptoForTestingOnly can be invoked by each Node independently (specifying the same seed, e.g. DefaultPseudoSeed)
// and generates the same set of keys for the whole system at each node, obviating the exchange of public keys.
func InsecureCryptoForTestingOnly(nodes []t.NodeID, ownID t.NodeID, seed int64) (Crypto, error) { //nolint:dupl

	// Create a new pseudorandom source from the given seed.
	randomness := insecureRNG.New(insecureRNG.NewSource(seed)) //nolint:gosec

	// Generate node keys.
	// All private keys except the own one will be discarded.
	nodePrivKeys, nodePubKeys, err := generateKeys(len(nodes), randomness)
	if err != nil {
		return nil, err
	}

	// Look up the own private key and create a CryptoImpl module instance that would sign with this key.
	var c *DefaultImpl
	for i, id := range nodes {
		if id == ownID {
			if c, err = NewDefaultImpl(nodePrivKeys[i]); err != nil {
				return nil, err
			}
		}
	}

	// Return error if own ID was not found in the given membership or CryptoImpl module instantiation failed
	if c == nil {
		if err != nil {
			// CryptoImpl module instantiation failed.
			return nil, err
		}

		// Own ID was not found and CryptoImpl module instantiation was not even attempted.
		return nil, es.Errorf("ownID (%v) not found among nodes", ownID)
	}

	// Populate the CryptoImpl module instance with the generated keys
	if err := registerPubKeys(c, nodes, nodePubKeys); err != nil {
		return nil, err
	}

	return c, nil
}

// generateKeys generates numKeys keys, using the given randomness source.
// returns private keys and public keys in two separate arrays, where privKeys[i] and pubKeys[i] represent one key pair.
func generateKeys(numKeys int, randomness io.Reader) (privKeys [][]byte, pubKeys [][]byte, err error) {

	// Initialize empty lists of keys.
	privKeys = make([][]byte, numKeys)
	pubKeys = make([][]byte, numKeys)

	// Generate key pairs.
	for i := 0; i < numKeys; i++ {
		if privKeys[i], pubKeys[i], err = GenerateKeyPair(randomness); err != nil {
			return nil, nil, err
		}
	}

	// Named output has already been set. Return.
	return
}

// regusterPubKeys populates a CryptoImpl module c with the given nodePubKeys.
// Each entry in nodes will be associated with the corresponding entry in nodePubKeys
// by calling c.RegisterNodeKey(nodePubKeys[i], nodes[i]) for 0 <= i < len(nodes).
// nodes and nodePubKeys must have the same length.
func registerPubKeys(
	c *DefaultImpl,
	nodes []t.NodeID,
	nodePubKeys [][]byte,
) error {

	// Populate CryptoImpl module with node keys.
	for keyIdx, nodeID := range nodes {
		if err := c.RegisterNodeKey(nodePubKeys[keyIdx], nodeID); err != nil {
			return err
		}
	}

	return nil
}
