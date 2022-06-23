package contextstore

import (
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Origin returns a ContextStoreOrigin protobuf containing the given id.
func Origin(itemID ItemID) *eventpb.ContextStoreOrigin {
	return &eventpb.ContextStoreOrigin{ItemID: itemID.Pb()}
}

// SignOrigin returns a SignOrigin protobuf containing moduleID and contextstore.Origin(itemID).
func SignOrigin(moduleID t.ModuleID, itemID ItemID) *eventpb.SignOrigin {
	return &eventpb.SignOrigin{
		Module: moduleID.Pb(),
		Type: &eventpb.SignOrigin_ContextStore{
			ContextStore: Origin(itemID),
		},
	}
}

// SigVerOrigin returns a SigVerOrigin protobuf containing moduleID and contextstore.Origin(itemID).
func SigVerOrigin(moduleID t.ModuleID, itemID ItemID) *eventpb.SigVerOrigin {
	return &eventpb.SigVerOrigin{
		Module: moduleID.Pb(),
		Type: &eventpb.SigVerOrigin_ContextStore{
			ContextStore: Origin(itemID),
		},
	}
}

// HashOrigin returns a HashOrigin protobuf containing moduleID and contextstore.Origin(itemID).
func HashOrigin(moduleID t.ModuleID, itemID ItemID) *eventpb.HashOrigin {
	return &eventpb.HashOrigin{
		Module: moduleID.Pb(),
		Type: &eventpb.HashOrigin_ContextStore{
			ContextStore: Origin(itemID),
		},
	}
}
