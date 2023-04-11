package isspbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/messagepb/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing net messages.

func UponISSMessageReceived[W types.ISSMessage_TypeWrapper[m], m any](m dsl.Module, handler func(from types1.NodeID, msg *m) error) {
	dsl1.UponMessageReceived[*types2.Message_Iss](m, func(from types1.NodeID, msg *types.ISSMessage) error {
		w, ok := msg.Type.(W)
		if !ok {
			return nil
		}

		return handler(from, w.Unwrap())
	})
}
