package iss

import (
	"crypto"
	"fmt"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/timer"
	t "github.com/filecoin-project/mir/pkg/types"
)

// DefaultModules takes a Modules object (as a value, not a pointer to it) and returns a pointer to a new Modules object
// with default ISS modules inserted in fields where no module has been specified.
func DefaultModules(orig modules.Modules) (modules.Modules, error) {

	// Create a new instance of Modules
	m := make(map[t.ModuleID]modules.Module)

	// Copy originally assigned modules
	for moduleID, module := range orig {
		m[moduleID] = module
	}

	// If no hasher module has been specified, use default SHA256 hasher.
	if m["hasher"] == nil {
		m["hasher"] = mirCrypto.NewHasher(crypto.SHA256)
	}

	if m["app"] == nil {
		return nil, fmt.Errorf("no default app implementation")
	}

	if m["crypto"] == nil {
		// TODO: Use default crypto once implemented and tested.
		return nil, fmt.Errorf("no default crypto implementation")
	}

	if m["iss"] == nil {
		return nil, fmt.Errorf("ISS protocol must be specified explicitly")
	}

	if m["timer"] == nil {
		m["timer"] = timer.New()
	}

	if m["net"] == nil {
		// TODO: Change this when a Net implementation exists.
		return nil, fmt.Errorf("no default Net implementation")
	}

	// If the WAL is not specified, no write-ahead log will be written
	// and the node will not be able to restart.
	if m["wal"] == nil {
		m["wal"] = modules.NullPassive{}
	}

	return m, nil
}
