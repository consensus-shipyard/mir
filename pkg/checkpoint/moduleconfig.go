package checkpoint

import t "github.com/filecoin-project/mir/pkg/types"

type ModuleConfig struct {
	Self t.ModuleID

	App    t.ModuleID
	Hasher t.ModuleID
	Crypto t.ModuleID
	Net    t.ModuleID
	Ord    t.ModuleID
}
