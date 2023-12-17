package dsl

import (
	dslpbtypes "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
)

// MirOrigin creates a dslpb.Origin protobuf.
func MirOrigin(contextID ContextID) *dslpbtypes.Origin {
	return &dslpbtypes.Origin{
		ContextID: contextID.Pb(),
	}
}
