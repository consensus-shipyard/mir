// Code generated by Mir codegen. DO NOT EDIT.

package communicationpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/blockchainpb/communicationpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Communication](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponNewBlock(m dsl.Module, handler func(block *types2.Block) error) {
	UponEvent[*types.Event_NewBlock](m, func(ev *types.NewBlock) error {
		return handler(ev.Block)
	})
}
