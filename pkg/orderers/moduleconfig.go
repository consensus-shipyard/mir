package orderers

import t "github.com/filecoin-project/mir/pkg/types"

type ModuleConfig struct {
	Self   t.ModuleID
	App    t.ModuleID
	Timer  t.ModuleID
	Hasher t.ModuleID
	Crypto t.ModuleID
	Wal    t.ModuleID
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
		Wal:    "wal",
		Net:    "net",
		Ord:    "iss",
		//Ava initially at runtime
	}
}
