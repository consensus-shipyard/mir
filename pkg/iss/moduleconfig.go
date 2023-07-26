package iss

import t "github.com/filecoin-project/mir/pkg/types"

// ModuleConfig contains the names of modules ISS depends on.
// The corresponding modules are expected by ISS to be stored under these keys by the Node.
type ModuleConfig struct {
	Self               t.ModuleID
	App                t.ModuleID
	Availability       t.ModuleID
	BatchDB            t.ModuleID
	Checkpoint         t.ModuleID
	ChkpValidator      t.ModuleID
	Net                t.ModuleID
	Ordering           t.ModuleID
	PPrepValidator     t.ModuleID
	PPrepValidatorChkp t.ModuleID
	Timer              t.ModuleID
}
