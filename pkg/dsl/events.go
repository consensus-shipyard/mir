package dsl

import (
	dslpbtypes "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
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
