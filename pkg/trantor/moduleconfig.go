package trantor

import (
	"github.com/filecoin-project/mir/pkg/availability/batchdb/fakebatchdb"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	"github.com/filecoin-project/mir/pkg/batchfetcher"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/checkpoint/chkpvalidator"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	ordererscommon "github.com/filecoin-project/mir/pkg/orderers/common"
	"github.com/filecoin-project/mir/pkg/orderers/common/pprepvalidator"
	t "github.com/filecoin-project/mir/pkg/types"
)

type ModuleConfig struct {
	App                t.ModuleID
	Availability       t.ModuleID
	BatchDB            t.ModuleID
	BatchFetcher       t.ModuleID
	Checkpointing      t.ModuleID
	ChkpValidator      t.ModuleID
	Crypto             t.ModuleID
	Hasher             t.ModuleID
	ISS                t.ModuleID // TODO: Rename this when Trantor is generalized to use other high-level protocols
	Mempool            t.ModuleID
	Net                t.ModuleID
	Null               t.ModuleID
	Ordering           t.ModuleID
	PPrepValidator     t.ModuleID
	PPrepValidatorChkp t.ModuleID
	Timer              t.ModuleID
}

func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{
		App:                "app",
		Availability:       "availability",
		BatchDB:            "batchdb",
		BatchFetcher:       "batchfetcher",
		Checkpointing:      "checkpoint",
		ChkpValidator:      "chkpvalidator",
		Crypto:             "crypto",
		Hasher:             "hasher",
		ISS:                "iss",
		Mempool:            "mempool",
		Net:                "net",
		Null:               "null",
		Ordering:           "ordering",
		PPrepValidator:     "pprepvalidator",
		PPrepValidatorChkp: "pprepvalidatorchkp",
		Timer:              "timer",
	}
}

func (mc ModuleConfig) ConfigureISS() iss.ModuleConfig {
	return iss.ModuleConfig{
		Self:               mc.ISS,
		App:                mc.BatchFetcher,
		Availability:       mc.Availability,
		BatchDB:            mc.BatchDB,
		Checkpoint:         mc.Checkpointing,
		ChkpValidator:      mc.ChkpValidator,
		Net:                mc.Net,
		Ordering:           mc.Ordering,
		PPrepValidator:     mc.PPrepValidator,
		PPrepValidatorChkp: mc.PPrepValidatorChkp,
		Timer:              mc.Timer,
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

func (mc ModuleConfig) ConfigureChkpValidator() chkpvalidator.ModuleConfig {
	return chkpvalidator.ModuleConfig{
		Self: mc.ChkpValidator,
	}
}

func (mc ModuleConfig) ConfigureOrdering() ordererscommon.ModuleConfig {
	return ordererscommon.ModuleConfig{
		Self:           mc.Ordering,
		App:            mc.BatchFetcher,
		Ava:            "", // Ava not initialized yet. It will be set at sub-module instantiation within the factory.
		Crypto:         mc.Crypto,
		Hasher:         mc.Hasher,
		Net:            mc.Net,
		Ord:            mc.ISS,
		PPrepValidator: mc.PPrepValidator,
		Timer:          mc.Timer,
	}
}

func (mc ModuleConfig) ConfigurePreprepareValidator() pprepvalidator.ModuleConfig {
	return pprepvalidator.ModuleConfig{
		Self: mc.PPrepValidator,
	}
}

func (mc ModuleConfig) ConfigurePreprepareValidatorChkp() pprepvalidator.ModuleConfig {
	return pprepvalidator.ModuleConfig{
		Self: mc.PPrepValidatorChkp,
	}
}

func (mc ModuleConfig) ConfigureSimpleMempool() simplemempool.ModuleConfig {
	return simplemempool.ModuleConfig{
		Self:   mc.Mempool,
		Hasher: mc.Hasher,
		Timer:  mc.Timer,
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
		Mempool:      mc.Mempool,
	}
}
