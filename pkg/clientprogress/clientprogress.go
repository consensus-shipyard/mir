package clientprogress

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ClientProgress tracks watermarks for all the clients.
type ClientProgress struct {
	ClientTrackers map[t.ClientID]*DeliveredReqs
	logger         logging.Logger
}

func NewClientProgress(logger logging.Logger) *ClientProgress {
	return &ClientProgress{
		ClientTrackers: make(map[t.ClientID]*DeliveredReqs),
		logger:         logger,
	}
}

func (cp *ClientProgress) Add(clID t.ClientID, reqNo t.ReqNo) bool {
	if _, ok := cp.ClientTrackers[clID]; !ok {
		cp.ClientTrackers[clID] = EmptyDeliveredReqs(logging.Decorate(cp.logger, "", "clID", clID))
	}
	return cp.ClientTrackers[clID].Add(reqNo)
}

func (cp *ClientProgress) GarbageCollect() map[t.ClientID]t.ReqNo {
	lowWMs := make(map[t.ClientID]t.ReqNo)
	for clientID, cwmt := range cp.ClientTrackers {
		lowWMs[clientID] = cwmt.GarbageCollect()
	}
	return lowWMs
}

func (cp *ClientProgress) Pb() *commonpb.ClientProgress {
	pb := make(map[string]*commonpb.DeliveredReqs)
	for clientID, clientTracker := range cp.ClientTrackers {
		pb[clientID.Pb()] = clientTracker.Pb()
	}
	return &commonpb.ClientProgress{Progress: pb}
}

func (cp *ClientProgress) LoadPb(pb *commonpb.ClientProgress) {
	cp.ClientTrackers = make(map[t.ClientID]*DeliveredReqs)
	for clientID, deliveredReqs := range pb.Progress {
		cp.ClientTrackers[t.ClientID(clientID)] = DeliveredReqsFromPb(
			deliveredReqs,
			logging.Decorate(cp.logger, "", "clID", clientID),
		)
	}
}

func FromPb(pb *commonpb.ClientProgress, logger logging.Logger) *ClientProgress {
	cp := NewClientProgress(logger)
	cp.LoadPb(pb)
	return cp
}
