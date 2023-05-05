package iss

import t "github.com/filecoin-project/mir/pkg/types"

// ModuleConfig contains the names of modules ISS depends on.
// The corresponding modules are expected by ISS to be stored under these keys by the Node.
type ModuleConfig struct {
	Self t.ModuleID

	App          t.ModuleID
	Availability t.ModuleID
	Checkpoint   t.ModuleID
	Net          t.ModuleID
	Ordering     t.ModuleID
	Timer        t.ModuleID
}

func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		Self: "iss",

		App:          "batchfetcher",
		Availability: "availability",
		Checkpoint:   "checkpointing",
		Net:          "net",
		Ordering:     "ordering",
		Timer:        "timer",
	}
}
