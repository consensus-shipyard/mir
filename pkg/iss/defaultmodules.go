package iss

import (
	"crypto"

	"github.com/pkg/errors"

	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/timer"
	t "github.com/filecoin-project/mir/pkg/types"
)

// DefaultModules takes a Modules object (as a value, not a pointer to it) and returns a pointer to a new Modules object
// with default ISS modules inserted in fields where no module has been specified.
func DefaultModules(orig modules.Modules, moduleConfig *ModuleConfig) (modules.Modules, error) {

	// Create a new instance of Modules
	m := make(map[t.ModuleID]modules.Module)

	// Copy originally assigned modules
	for moduleID, module := range orig {
		m[moduleID] = module
	}

	// If no hasher module has been specified, use default SHA256 hasher.
	if m[moduleConfig.Hasher] == nil {
		m[moduleConfig.Hasher] = mirCrypto.NewHasher(crypto.SHA256)
	}

	if m[moduleConfig.App] == nil {
		return nil, errors.New("no default ISS app implementation")
	}

	if m[moduleConfig.Crypto] == nil {
		// TODO: Use default crypto once implemented and tested.
		return nil, errors.New("no default crypto implementation")
	}

	if m[moduleConfig.Self] == nil {
		return nil, errors.New("ISS protocol must be specified explicitly")
	}

	if m[moduleConfig.Timer] == nil {
		m[moduleConfig.Timer] = timer.New()
	}

	if m[moduleConfig.Net] == nil {
		// TODO: Change this when a Net implementation exists.
		return nil, errors.New("no default Net implementation")
	}

	return m, nil
}
