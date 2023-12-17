package stdevents

import t "github.com/filecoin-project/mir/pkg/types"

type mirEvent struct {
	SrcModule  t.ModuleID
	DestModule t.ModuleID
}

func (e *mirEvent) Src() t.ModuleID {
	return e.SrcModule
}

func (e *mirEvent) Dest() t.ModuleID {
	return e.DestModule
}
