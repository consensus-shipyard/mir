package iss

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/clients"
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/reqstore"
	"github.com/filecoin-project/mir/pkg/timer"
	t "github.com/filecoin-project/mir/pkg/types"
)

// DefaultModules takes a Modules object and returns a new Modules object
// with default ISS modules inserted under IDs where no module has been specified.
func DefaultModules(m modules.Modules) (modules.Modules, error) {

	// Copy assigned modules
	withDefaults := make(map[t.ModuleID]modules.Module)
	for moduleID, module := range m {
		withDefaults[moduleID] = module
	}

	if withDefaults["net"] == nil {
		// TODO: Change this when a Net implementation exists.
		return nil, fmt.Errorf("no default Net implementation")
	}

	if withDefaults["hasher"] == nil {
		withDefaults["hasher"] = modules.PassiveModule(&crypto.SHA256Hasher{})
	}

	if withDefaults["app"] == nil {
		return nil, fmt.Errorf("no default App implementation")
	}

	if withDefaults["clientTracker"] == nil {
		// TODO: Change this to the real default client tracker once implemented.
		//       Also, make the "iss" default protocol module more explicit.
		withDefaults["clientTracker"] = clients.SigningTracker("iss", nil)
	}

	if withDefaults["requestStore"] == nil {
		withDefaults["requestStore"] = reqstore.NewVolatileRequestStore()
	}

	if withDefaults["iss"] == nil {
		// TODO: Use default protocol once implemented.
		return nil, fmt.Errorf("no default protocol implementation")
	}

	if withDefaults["crypto"] == nil {
		// TODO: Use default crypto once implemented and tested.
		return nil, fmt.Errorf("no default crypto implementation")
	}

	if withDefaults["timer"] == nil {
		withDefaults["timer"] = modules.ActiveModule(&timer.Timer{})
	}

	// The Interceptor can stay nil, in which case Events will simply not be intercepted.

	// The WAL can stay nil, in which case no write-ahead log will be written
	// and the node will not be able to restart.

	return withDefaults, nil
}
