package availabilitypbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/availabilitypb/events"
	types2 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func CertVerified(m dsl.Module, next []*types.Event, destModule types1.ModuleID, valid bool, err string, origin *types2.VerifyCertOrigin) {
	dsl.EmitMirEvent(m, events.CertVerified(next, destModule, valid, err, origin))
}
