package common

import t "github.com/filecoin-project/mir/pkg/types"

type ModuleConfig struct {
	Self t.ModuleID

	App            t.ModuleID
	Ava            t.ModuleID
	PPrepValidator t.ModuleID
	Crypto         t.ModuleID
	Hasher         t.ModuleID
	Net            t.ModuleID
	Ord            t.ModuleID
	Timer          t.ModuleID
}
