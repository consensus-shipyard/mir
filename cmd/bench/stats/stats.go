package stats

import trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"

type StatsTracker interface {
	Submit(tx *trantorpbtypes.Transaction)
	Deliver(tx *trantorpbtypes.Transaction)
}
