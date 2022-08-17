package availabilitypbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Availability](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponCertVerified(m dsl.Module, handler func(valid bool, err string, origin *types.VerifyCertOrigin) error) {
	UponEvent[*types.Event_CertVerified](m, func(ev *types.CertVerified) error {
		return handler(ev.Valid, ev.Err, ev.Origin)
	})
}
