package contextstore

import (
	contextstorepbtypes "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Origin returns a contextstorepb.Origin protobuf containing the given id.
func Origin(itemID ItemID) *contextstorepbtypes.Origin {
	return &contextstorepbtypes.Origin{ItemID: itemID.Pb()}
}

// SignOrigin returns a SignOrigin protobuf containing moduleID and contextstore.Origin(itemID).
func SignOrigin(moduleID t.ModuleID, itemID ItemID) *cryptopbtypes.SignOrigin {
	return &cryptopbtypes.SignOrigin{
		Module: moduleID,
		Type: &cryptopbtypes.SignOrigin_ContextStore{
			ContextStore: Origin(itemID),
		},
	}
}

// SigVerOrigin returns a SigVerOrigin protobuf containing moduleID and contextstore.Origin(itemID).
func SigVerOrigin(moduleID t.ModuleID, itemID ItemID) *cryptopbtypes.SigVerOrigin {
	return &cryptopbtypes.SigVerOrigin{
		Module: moduleID,
		Type: &cryptopbtypes.SigVerOrigin_ContextStore{
			ContextStore: Origin(itemID),
		},
	}
}

// HashOrigin returns a HashOrigin protobuf containing moduleID and contextstore.Origin(itemID).
func HashOrigin(moduleID t.ModuleID, itemID ItemID) *hasherpbtypes.HashOrigin {
	return &hasherpbtypes.HashOrigin{
		Module: moduleID,
		Type: &hasherpbtypes.HashOrigin_ContextStore{
			ContextStore: Origin(itemID),
		},
	}
}
