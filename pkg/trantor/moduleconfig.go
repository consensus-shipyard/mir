package trantor

import t "github.com/filecoin-project/mir/pkg/types"

type ModuleConfig struct {
	App           t.ModuleID
	Availability  t.ModuleID
	BatchDB       t.ModuleID
	BatchFetcher  t.ModuleID
	Checkpointing t.ModuleID
	Crypto        t.ModuleID
	Hasher        t.ModuleID
	ISS           t.ModuleID // TODO: Rename this when trantor is generalized to use other high-level protocols
	Mempool       t.ModuleID
	Net           t.ModuleID
	Null          t.ModuleID
	Ordering      t.ModuleID
	Timer         t.ModuleID
}

func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{
		App:           "app",
		Availability:  "availability",
		BatchDB:       "batchdb",
		BatchFetcher:  "batchfetcher",
		Checkpointing: "checkpoint",
		Crypto:        "crypto",
		Hasher:        "hasher",
		ISS:           "iss",
		Mempool:       "mempool",
		Net:           "net",
		Null:          "null",
		Ordering:      "ordering",
		Timer:         "timer",
	}
}
