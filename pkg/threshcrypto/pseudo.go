/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package threshcrypto

import (
	"crypto/cipher"
	"fmt"
	prand "math/rand"

	"github.com/drand/kyber/util/random"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

var (
	// DefaultPseudoSeed is an arbitrary number that the nodes can use as a seed when instantiating its MirModule module.
	// This is not secure, but helps during testing, as it obviates the exchange of public keys among nodes.
	DefaultPseudoSeed int64 = 12345
)

func pseudorandomStream(seed int64) cipher.Stream {
	pseudorand := prand.New(prand.NewSource(seed)) //nolint:gosec
	return random.New(pseudorand)
}

// NodePseudo returns a ThreshCryptoImpl module to be used by a Node, generating new keys in a pseudo-random manner.
// It is initialized and populated deterministically, based on a given configuration and a random seed.
// NodePseudo is not secure.
// Intended for testing purposes and assuming a static membership known to all nodes,
// NodePseudo can be invoked by each Node independently (specifying the same seed, e.g. DefaultPseudoSeed)
// and generates the same set of keys for the whole system at each node, obviating the exchange of public keys.
func TBLSPseudo(nodes []t.NodeID, threshold int, ownID t.NodeID, seed int64) (ThreshCrypto, error) { //nolint:dupl
	// Create a new pseudorandom source from the given seed.
	randomness := pseudorandomStream(seed)

	var idx int
	var ok bool
	if ok, idx = sliceutil.IndexOf(nodes, ownID); !ok {
		return nil, fmt.Errorf("own node ID not in node list")
	}

	if tcInstances, err := TBLS12381Keygen(threshold, len(nodes), randomness); err != nil {
		return nil, err
	} else {
		return tcInstances[idx], nil
	}
}
