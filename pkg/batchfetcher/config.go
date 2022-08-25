package batchfetcher

import t "github.com/filecoin-project/mir/pkg/types"

// ModuleConfig determines the IDs of the modules the batch fetcher interacts with.
type ModuleConfig struct {
	Self         t.ModuleID // Own ID.
	Availability t.ModuleID // ID of the factory module containing the availability modules.
	Destination  t.ModuleID // ID of the module to deliver the produced event stream to (usually the application).
}

func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		Self:         "batchFetcher",
		Availability: "availability",
		Destination:  "app",
	}
}
