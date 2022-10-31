package clientprogress

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ClientProgress tracks watermarks for all the clients.
type ClientProgress struct {
	clientTrackers map[t.ClientID]*DeliveredReqs
	logger         logging.Logger
}

func NewClientProgress(logger logging.Logger) *ClientProgress {
	return &ClientProgress{
		clientTrackers: make(map[t.ClientID]*DeliveredReqs),
		logger:         logger,
	}
}

func (wmt *ClientProgress) Add(clID t.ClientID, reqNo t.ReqNo) bool {
	if _, ok := wmt.clientTrackers[clID]; !ok {
		wmt.clientTrackers[clID] = EmptyDeliveredReqs(wmt.logger)
	}
	return wmt.clientTrackers[clID].Add(reqNo)
}

func (wmt *ClientProgress) GarbageCollect() map[t.ClientID]t.ReqNo {
	lowWMs := make(map[t.ClientID]t.ReqNo)
	for clientID, cwmt := range wmt.clientTrackers {
		lowWMs[clientID] = cwmt.GarbageCollect()
	}
	return lowWMs
}

func (wmt *ClientProgress) Pb() *commonpb.ClientProgress {
	cp := make(map[string]*commonpb.DeliveredReqs)
	for clientID, clientTracker := range wmt.clientTrackers {
		cp[clientID.Pb()] = clientTracker.Pb()
	}
	return &commonpb.ClientProgress{Progress: cp}
}

func (wmt *ClientProgress) LoadPb(cp *commonpb.ClientProgress) {
	wmt.clientTrackers = make(map[t.ClientID]*DeliveredReqs)
	for clientID, deliveredReqs := range cp.Progress {
		wmt.clientTrackers[t.ClientID(clientID)] = DeliveredReqsFromPb(deliveredReqs, wmt.logger)
	}
}

func FromPb(pb *commonpb.ClientProgress, logger logging.Logger) *ClientProgress {
	cp := NewClientProgress(logger)
	for clientID, deliveredReqs := range pb.Progress {
		cp.clientTrackers[t.ClientID(clientID)] = DeliveredReqsFromPb(deliveredReqs, logger)
	}
	return cp
}
