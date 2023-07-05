package dsl

import (
	dslpbtypes "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
)

// Origin creates a dslpbtypes.Origin protobuf.
func Origin(contextID ContextID) *dslpbtypes.Origin {
	return &dslpbtypes.Origin{
		ContextID: contextID.Pb(),
	}
}

// MirOrigin creates a dslpbtypes.Origin protobuf.
func MirOrigin(contextID ContextID) *dslpbtypes.Origin {
	return &dslpbtypes.Origin{
		ContextID: contextID.Pb(),
	}
}

// UponInit invokes handler when the module is initialized.
func UponInit(m Module, handler func() error) {
	UponEvent[*eventpbtypes.Event_Init](m, func(ev *eventpbtypes.Init) error {
		return handler()
	})
}
