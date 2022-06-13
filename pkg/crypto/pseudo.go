/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package crypto

import (
	"fmt"
	"io"
	prand "math/rand"

	t "github.com/filecoin-project/mir/pkg/types"
)

var (
	// DefaultPseudoSeed is an arbitrary number that the nodes can use as a seed when instantiating its Crypto module.
	// This is not secure, but helps during testing, as it obviates the exchange of public keys among nodes.
	DefaultPseudoSeed int64 = 12345
)

// NodePseudo returns a CryptoImpl module to be used by a Node, generating new keys in a pseudo-random manner.
// It is initialized and populated deterministically, based on a given configuration and a random seed.
// NodePseudo is not secure.
// Intended for testing purposes and assuming a static membership known to all nodes,
// NodePseudo can be invoked by each Node independently (specifying the same seed, e.g. DefaultPseudoSeed)
// and generates the same set of keys for the whole system at each node, obviating the exchange of public keys.
func NodePseudo(nodes []t.NodeID, clients []t.ClientID, ownID t.NodeID, seed int64) (Impl, error) { //nolint:dupl

	// Create a new pseudorandom source from the given seed.
	randomness := prand.New(prand.NewSource(seed)) //nolint:gosec

	// Generate node keys.
	// All private keys except the own one will be discarded.
	nodePrivKeys, nodePubKeys, err := generateKeys(len(nodes), randomness)
	if err != nil {
		return nil, err
	}

	// Generate client keys.
	// All client private keys are discarded.
	_, clientPubKeys, err := generateKeys(len(clients), randomness)
	if err != nil {
		return nil, err
	}

	// Look up the own private key and create a CryptoImpl module instance that would sign with this key.
	var c Impl
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
		return nil, fmt.Errorf("ownID (%v) not found among nodes", ownID)
	}

	// Populate the CryptoImpl module instance with the generated keys
	if err := registerPubKeys(c, nodes, nodePubKeys, clients, clientPubKeys); err != nil {
		return nil, err
	}

	return c, nil
}

// ClientPseudo behaves the same as NodePseudo, except that it returns a CryptoImpl module intended for use by the client.
// The returned CryptoImpl module will use the private key associated with client ownID for signing.
func ClientPseudo(nodes []t.NodeID, clients []t.ClientID, ownID t.ClientID, seed int64) (Impl, error) { //nolint:dupl

	// Create a new pseudorandom source from the given seed.
	randomness := prand.New(prand.NewSource(seed)) //nolint:gosec

	// Generate node keys.
	// All node private keys are discarded.
	_, nodePubKeys, err := generateKeys(len(nodes), randomness)
	if err != nil {
		return nil, err
	}

	// Generate client keys.
	// All client private keys except the own one will be discarded.
	clientPrivKeys, clientPubKeys, err := generateKeys(len(clients), randomness)
	if err != nil {
		return nil, err
	}

	// Look up the own private key and create a CryptoImpl module instance that would sign with this key.
	var c Impl
	for i, id := range clients {
		if id == ownID {
			if c, err = NewDefaultImpl(clientPrivKeys[i]); err != nil {
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
		return nil, fmt.Errorf("ownID (%v) not found among clients", ownID)
	}

	// Populate the CryptoImpl module instance with the generated keys
	if err := registerPubKeys(c, nodes, nodePubKeys, clients, clientPubKeys); err != nil {
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

// regusterPubKeys populates a CryptoImpl module c with the given nodePubKeys and clientPubKeys.
// Each entry in nodes will be associated with the corresponding entry in nodePubKeys
// by calling c.RegisterNodeKey(nodePubKeys[i], nodes[i]) for 0 <= i < len(nodes).
// nodes and nodePubKeys must have the same length.
// The analogous happens for client keys, using c.RegisterClientKey.
func registerPubKeys(
	c Impl,
	nodes []t.NodeID,
	nodePubKeys [][]byte,
	clients []t.ClientID,
	clientPubKeys [][]byte,
) error {

	// Populate CryptoImpl module with node keys.
	for keyIdx, nodeID := range nodes {
		if err := c.RegisterNodeKey(nodePubKeys[keyIdx], nodeID); err != nil {
			return err
		}
	}

	// Populate CryptoImpl module with client keys.
	for keyIdx, clientID := range clients {
		if err := c.RegisterClientKey(clientPubKeys[keyIdx], clientID); err != nil {
			return err
		}
	}

	return nil
}
