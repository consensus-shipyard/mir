package availabilitypbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/availabilitypb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func CertVerified(m dsl.Module, destModule types.ModuleID, valid bool, err string, origin *types1.VerifyCertOrigin) {
	dsl.EmitMirEvent(m, events.CertVerified(destModule, valid, err, origin))
}
