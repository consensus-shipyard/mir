package orderers

import t "github.com/filecoin-project/mir/pkg/types"

type ModuleConfig struct {
	Self t.ModuleID

	App    t.ModuleID
	Ava    t.ModuleID
	Crypto t.ModuleID
	Hasher t.ModuleID
	Net    t.ModuleID
	Ord    t.ModuleID
	Timer  t.ModuleID
}

func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		Self: "ordering",

		App: "batchfetcher",
		// Ava not initialized by default. Must be set at module instantiation.
		Crypto: "crypto",
		Hasher: "hasher",
		Net:    "net",
		Ord:    "iss",
		Timer:  "timer",
	}
}
