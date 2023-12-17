package threshcrypto

import (
	"crypto/cipher"
	prand "math/rand"

	"github.com/drand/kyber/util/random"
	t "github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
	"golang.org/x/exp/slices"
)

var (
	// DefaultPseudoSeed is an arbitrary number that the nodes can use as a seed when instantiating their MirModule modules.
	// This is not secure, but helps during testing, as it obviates the exchange of public keys among nodes.
	DefaultPseudoSeed int64 = 12345
)

// pseudorandomStream creates a deterministic cipher.Stream from a seed
func pseudorandomStream(seed int64) cipher.Stream {
	pseudorand := prand.New(prand.NewSource(seed)) //nolint:gosec
	return random.New(pseudorand)
}

// TBLSPseudo returns a ThreshCryptoImpl module to be used by a Node, generating new keys in a pseudo-random manner.
// It is initialized and populated deterministically, based on a given configuration and a random seed.
// NodePseudo is not secure.
// Intended for testing purposes and assuming a static membership known to all nodes,
// NodePseudo can be invoked by each Node independently (specifying the same seed, e.g. DefaultPseudoSeed)
// and generates the same set of keys for the whole system at each node, obviating the exchange of public keys.
func TBLSPseudo(nodes []t.NodeID, threshold int, ownID t.NodeID, seed int64) (ThreshCrypto, error) {
	// Create a new pseudorandom source from the given seed.
	randomness := pseudorandomStream(seed)

	idx := slices.Index(nodes, ownID)
	if idx == -1 {
		return nil, es.Errorf("own node ID not in node list")
	}

	tcInstances := TBLS12381Keygen(threshold, nodes, randomness)

	return tcInstances[idx], nil
}
