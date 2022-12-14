package orderers

import t "github.com/filecoin-project/mir/pkg/types"

type ModuleConfig struct {
	Self   t.ModuleID
	App    t.ModuleID
	Timer  t.ModuleID
	Hasher t.ModuleID
	Crypto t.ModuleID
	Net    t.ModuleID
	Ord    t.ModuleID
	Ava    t.ModuleID
}

func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		Self:   "ordering",
		Timer:  "timer",
		App:    "batchfetcher",
		Hasher: "hasher",
		Crypto: "crypto",
		Net:    "net",
		Ord:    "iss",
		//Ava not initialized by default. Must be set at module instantiation.
	}
}
