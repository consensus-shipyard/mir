// Code generated by Mir codegen. DO NOT EDIT.

package interceptorpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	blockchainpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	types "github.com/filecoin-project/mir/pkg/pb/blockchainpb/interceptorpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Bcinterceptor](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponTreeUpdate(m dsl.Module, handler func(tree *blockchainpb.Blocktree, headId uint64) error) {
	UponEvent[*types.Event_TreeUpdate](m, func(ev *types.TreeUpdate) error {
		return handler(ev.Tree, ev.HeadId)
	})
}

func UponNewOrphan(m dsl.Module, handler func(orphan *blockchainpb.Block) error) {
	UponEvent[*types.Event_NewOrphan](m, func(ev *types.NewOrphan) error {
		return handler(ev.Orphan)
	})
}
