package trantor

import (
	"github.com/filecoin-project/mir/pkg/eventmangler"
	"github.com/filecoin-project/mir/pkg/types"
)

// PerturbMessages configures the SMR system to randomly drop and delay some of the messages sent over the network.
// Useful for debugging and stress-testing.
// The params argument defines parameters of the perturbation, such as how many messages should be dropped
// and how the remaining messages should be delayed.
func PerturbMessages(params *eventmangler.ModuleParams, moduleID types.ModuleID, sys *System) error {

	// Create event mangler perturbing (dropping and delaying) events.
	messageMangler, err := eventmangler.NewModule(
		eventmangler.ModuleConfig{Self: moduleID, Dest: "truenet", Timer: "timer"},
		params,
	)
	if err != nil {
		return err
	}

	// Intercept all events (in this case SendMessage events) directed to the "net" module by the mangler
	// And change the actual transport module ID to "truenet", where the mangler forwards the surviving messages.
	sys.modules[moduleID] = messageMangler
	sys.modules["truenet"] = sys.transport
	return nil
}
