package clientprogress

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/trantorpb"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

// ClientProgress tracks watermarks for all the clients.
type ClientProgress struct {
	ClientTrackers map[tt.ClientID]*DeliveredTXs
	logger         logging.Logger
}

func NewClientProgress(logger logging.Logger) *ClientProgress {
	return &ClientProgress{
		ClientTrackers: make(map[tt.ClientID]*DeliveredTXs),
		logger:         logger,
	}
}

func (cp *ClientProgress) Contains(clID tt.ClientID, txNo tt.TxNo) bool {
	if _, ok := cp.ClientTrackers[clID]; !ok {
		return false
	}
	return cp.ClientTrackers[clID].Contains(txNo)
}

func (cp *ClientProgress) Add(clID tt.ClientID, txNo tt.TxNo) bool {
	if _, ok := cp.ClientTrackers[clID]; !ok {
		cp.ClientTrackers[clID] = EmptyDeliveredTXs(logging.Decorate(cp.logger, "", "clID", clID))
	}
	return cp.ClientTrackers[clID].Add(txNo)
}

func (cp *ClientProgress) GarbageCollect() map[tt.ClientID]tt.TxNo {
	lowWMs := make(map[tt.ClientID]tt.TxNo)
	for clientID, cwmt := range cp.ClientTrackers {
		lowWMs[clientID] = cwmt.GarbageCollect()
	}
	return lowWMs
}

func (cp *ClientProgress) Pb() *trantorpb.ClientProgress {
	pb := make(map[string]*trantorpb.DeliveredTXs)
	for clientID, clientTracker := range cp.ClientTrackers {
		pb[clientID.Pb()] = clientTracker.Pb()
	}
	return &trantorpb.ClientProgress{Progress: pb}
}

func (cp *ClientProgress) LoadPb(pb *trantorpb.ClientProgress) {
	cp.ClientTrackers = make(map[tt.ClientID]*DeliveredTXs)
	for clientID, deliveredTXs := range pb.Progress {
		cp.ClientTrackers[tt.ClientID(clientID)] = DeliveredTXsFromPb(
			deliveredTXs,
			logging.Decorate(cp.logger, "", "clID", clientID),
		)
	}
}

func FromPb(pb *trantorpb.ClientProgress, logger logging.Logger) *ClientProgress {
	cp := NewClientProgress(logger)
	cp.LoadPb(pb)
	return cp
}
