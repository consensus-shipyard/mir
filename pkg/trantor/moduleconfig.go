package trantor

import (
	"github.com/filecoin-project/mir/pkg/availability/batchdb/fakebatchdb"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	"github.com/filecoin-project/mir/pkg/batchfetcher"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	ordererscommon "github.com/filecoin-project/mir/pkg/orderers/common"
	t "github.com/filecoin-project/mir/pkg/types"
)

type ModuleConfig struct {
	App           t.ModuleID
	Availability  t.ModuleID
	BatchDB       t.ModuleID
	BatchFetcher  t.ModuleID
	Checkpointing t.ModuleID
	Crypto        t.ModuleID
	Hasher        t.ModuleID
	ISS           t.ModuleID // TODO: Rename this when Trantor is generalized to use other high-level protocols
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

func (mc ModuleConfig) ConfigureISS() iss.ModuleConfig {
	return iss.ModuleConfig{
		Self:         mc.ISS,
		App:          mc.BatchFetcher,
		Availability: mc.Availability,
		Checkpoint:   mc.Checkpointing,
		Net:          mc.Net,
		Ordering:     mc.Ordering,
		Timer:        mc.Timer,
	}
}
func (mc ModuleConfig) ConfigureCheckpointing() checkpoint.ModuleConfig {
	return checkpoint.ModuleConfig{
		Self:   mc.Checkpointing,
		App:    mc.BatchFetcher,
		Crypto: mc.Crypto,
		Hasher: mc.Hasher,
		Net:    mc.Net,
		Ord:    mc.ISS,
	}
}

func (mc ModuleConfig) ConfigureOrdering() ordererscommon.ModuleConfig {
	return ordererscommon.ModuleConfig{
		Self:   mc.Ordering,
		App:    mc.BatchFetcher,
		Ava:    "", // Ava not initialized yet. It will be set at sub-module instantiation within the factory.
		Crypto: mc.Crypto,
		Hasher: mc.Hasher,
		Net:    mc.Net,
		Ord:    mc.ISS,
		Timer:  mc.Timer,
	}
}

func (mc ModuleConfig) ConfigureSimpleMempool() simplemempool.ModuleConfig {
	return simplemempool.ModuleConfig{
		Self:   mc.Mempool,
		Hasher: mc.Hasher,
	}
}

func (mc ModuleConfig) ConfigureFakeBatchDB() fakebatchdb.ModuleConfig {
	return fakebatchdb.ModuleConfig{
		Self: mc.BatchDB,
	}
}

func (mc ModuleConfig) ConfigureMultisigCollector() multisigcollector.ModuleConfig {
	return multisigcollector.ModuleConfig{
		Self:    mc.Availability,
		Mempool: mc.Mempool,
		BatchDB: mc.BatchDB,
		Net:     mc.Net,
		Crypto:  mc.Crypto,
	}
}

func (mc ModuleConfig) ConfigureBatchFetcher() batchfetcher.ModuleConfig {
	return batchfetcher.ModuleConfig{
		Self:         mc.BatchFetcher,
		Availability: mc.Availability,
		Checkpoint:   mc.Checkpointing,
		Destination:  mc.App,
	}
}
