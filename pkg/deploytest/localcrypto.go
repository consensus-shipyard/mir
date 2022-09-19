package deploytest

import (
	"fmt"

	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"
)

type LocalCryptoSystem interface {
	Crypto(id t.NodeID) mirCrypto.Crypto
	Module(id t.NodeID) modules.Module
}

type localPseudoCryptoSystem struct {
	nodeIDs []t.NodeID
}

// NewLocalCryptoSystem creates an instance of LocalCryptoSystem suitable for tests.
// In the current implementation, cryptoType can only be "pseudo".
func NewLocalCryptoSystem(cryptoType string, nodeIDs []t.NodeID, logger logging.Logger) LocalCryptoSystem {
	return &localPseudoCryptoSystem{nodeIDs}
}

func (cs *localPseudoCryptoSystem) Crypto(id t.NodeID) mirCrypto.Crypto {
	cryptoImpl, err := mirCrypto.NodePseudo(cs.nodeIDs, id, mirCrypto.DefaultPseudoSeed)
	if err != nil {
		panic(fmt.Sprintf("error creating crypto module: %v", err))
	}
	return cryptoImpl
}

func (cs *localPseudoCryptoSystem) Module(id t.NodeID) modules.Module {
	return mirCrypto.New(cs.Crypto(id))
}
