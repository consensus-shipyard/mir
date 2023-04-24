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
	nodeIDs []t.NodeID
}

// NewLocalCryptoSystem creates an instance of LocalCryptoSystem suitable for tests.
// In the current implementation, cryptoType can only be "pseudo".
func NewLocalCryptoSystem(_ string, nodeIDs []t.NodeID, _ logging.Logger) LocalCryptoSystem {
	return &localPseudoCryptoSystem{nodeIDs}
}

func (cs *localPseudoCryptoSystem) Crypto(id t.NodeID) (mirCrypto.Crypto, error) {
	return mirCrypto.InsecureCryptoForTestingOnly(cs.nodeIDs, id, mirCrypto.DefaultPseudoSeed)
}

func (cs *localPseudoCryptoSystem) Module(id t.NodeID) (modules.Module, error) {
	c, err := cs.Crypto(id)
	if err != nil {
		return nil, err
	}
	return mirCrypto.New(c), nil
}
