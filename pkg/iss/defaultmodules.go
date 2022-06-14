package iss

import (
	"crypto"
	"fmt"
	"github.com/filecoin-project/mir/pkg/clients"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/reqstore"
	"github.com/filecoin-project/mir/pkg/timer"
	t "github.com/filecoin-project/mir/pkg/types"
)

// DefaultModules takes a Modules object (as a value, not a pointer to it) and returns a pointer to a new Modules object
// with default ISS modules inserted in fields where no module has been specified.
func DefaultModules(m modules.Modules) (*modules.Modules, error) {
	if m.Net == nil {
		// TODO: Change this when a Net implementation exists.
		panic("no default Net implementation")
	}

	if m.Timer == nil {
		m.Timer = &timer.Timer{}
	}

	// The Interceptor can stay nil, in which case Events will simply not be intercepted.

	// Copy assigned generic modules
	if m.GenericModules != nil {
		gm := m.GenericModules
		m.GenericModules = make(map[t.ModuleID]modules.Module)
		for moduleID, module := range gm {
			m.GenericModules[moduleID] = module
		}
	} else {
		m.GenericModules = make(map[t.ModuleID]modules.Module)
	}

	// If no hasher module has been specified, use default SHA256 hasher.
	if m.GenericModules["hasher"] == nil {
		m.GenericModules["hasher"] = mirCrypto.NewHasher(crypto.SHA256)
	}

	if m.GenericModules["app"] == nil {
		return nil, fmt.Errorf("no default app implementation")
	}

	if m.GenericModules["crypto"] == nil {
		// TODO: Use default crypto once implemented and tested.
		return nil, fmt.Errorf("no default crypto implementation")
	}

	if m.GenericModules["clientTracker"] == nil {
		m.GenericModules["clientTracker"] = clients.SigningTracker("iss", nil)
	}

	if m.GenericModules["requestStore"] == nil {
		m.GenericModules["requestStore"] = reqstore.NewVolatileRequestStore()
	}

	if m.GenericModules["iss"] == nil {
		return nil, fmt.Errorf("ISS protocol must be specified explicitly")
	}

	// The WAL can stay nil, in which case no write-ahead log will be written
	// and the node will not be able to restart.

	return &m, nil
}
