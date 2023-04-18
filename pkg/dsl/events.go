package dsl

import (
	dslpbtypes "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

// Origin creates a dslpb.Origin protobuf.
func Origin(contextID ContextID) *dslpbtypes.Origin {
	return &dslpbtypes.Origin{
		ContextID: contextID.Pb(),
	}
}

// MirOrigin creates a dslpb.Origin protobuf.
func MirOrigin(contextID ContextID) *dslpbtypes.Origin {
	return &dslpbtypes.Origin{
		ContextID: contextID.Pb(),
	}
}

// UponInit invokes handler when the module is initialized.
func UponInit(m Module, handler func() error) {
	UponEvent[*eventpb.Event_Init](m, func(ev *eventpb.Init) error {
		return handler()
	})
}
