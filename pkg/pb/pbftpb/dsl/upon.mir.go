package pbftpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types2 "github.com/filecoin-project/mir/pkg/orderers/types"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/ordererpb/dsl"
	types1 "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl1.UponEvent[*types1.Event_Pbft](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponProposeTimeout(m dsl.Module, handler func(proposeTimeout uint64) error) {
	UponEvent[*types.Event_ProposeTimeout](m, func(ev *types.ProposeTimeout) error {
		return handler(ev.ProposeTimeout)
	})
}

func UponViewChangeSNTimeout(m dsl.Module, handler func(view types2.ViewNr, numCommitted uint64) error) {
	UponEvent[*types.Event_ViewChangeSnTimeout](m, func(ev *types.ViewChangeSNTimeout) error {
		return handler(ev.View, ev.NumCommitted)
	})
}

func UponViewChangeSegTimeout(m dsl.Module, handler func(viewChangeSegTimeout uint64) error) {
	UponEvent[*types.Event_ViewChangeSegTimeout](m, func(ev *types.ViewChangeSegTimeout) error {
		return handler(ev.ViewChangeSegTimeout)
	})
}
