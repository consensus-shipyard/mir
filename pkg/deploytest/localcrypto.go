package deploytest

import (
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"
)

type LocalCryptoSystem interface {
	Crypto(id t.NodeID) (mirCrypto.Crypto, error)
	Module(id t.NodeID) (modules.Module, error)
}

type localPseudoCryptoSystem struct {
	nodeIDs       []t.NodeID
	localKeyPairs mirCrypto.KeyPairs
}

// NewLocalCryptoSystem creates an instance of LocalCryptoSystem suitable for tests.
// In the current implementation, cryptoType can only be "pseudo".
// TODO: Think about merging all the code in this file with the pseudo crypto implementation.
func NewLocalCryptoSystem(_ string, nodeIDs []t.NodeID, _ logging.Logger) (LocalCryptoSystem, error) {

	// Generate keys once. All crypto modules derived from this crypto system will use those keys.
	keyPairs, err := mirCrypto.GenerateKeys(len(nodeIDs), mirCrypto.DefaultPseudoSeed)
	if err != nil {
		return nil, err
	}
	return &localPseudoCryptoSystem{nodeIDs, keyPairs}, nil
}

func (cs *localPseudoCryptoSystem) Crypto(id t.NodeID) (mirCrypto.Crypto, error) {
	return mirCrypto.InsecureCryptoForTestingOnly(cs.nodeIDs, id, &cs.localKeyPairs)
}

func (cs *localPseudoCryptoSystem) Module(id t.NodeID) (modules.Module, error) {
	c, err := cs.Crypto(id)
	if err != nil {
		return nil, err
	}
	return mirCrypto.New(c), nil
}
