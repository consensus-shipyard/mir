package clientprogress

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/trantorpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
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
	lowWMs := make(map[tt.ClientID]tt.TxNo, len(cp.ClientTrackers))
	for clientID, cwmt := range cp.ClientTrackers {
		lowWMs[clientID] = cwmt.GarbageCollect()
	}
	return lowWMs
}

func (cp *ClientProgress) Pb() *trantorpb.ClientProgress {
	pb := make(map[string]*trantorpb.DeliveredTXs, len(cp.ClientTrackers))
	for clientID, clientTracker := range cp.ClientTrackers {
		pb[clientID.Pb()] = clientTracker.Pb()
	}
	return &trantorpb.ClientProgress{Progress: pb}
}

func (cp *ClientProgress) LoadPb(pb *trantorpb.ClientProgress) {
	cp.ClientTrackers = make(map[tt.ClientID]*DeliveredTXs, len(pb.Progress))
	for clientID, deliveredTXs := range pb.Progress {
		cp.ClientTrackers[tt.ClientID(clientID)] = DeliveredTXsFromPb(
			deliveredTXs,
			logging.Decorate(cp.logger, "", "clID", clientID),
		)
	}
}

func (cp *ClientProgress) LoadDslStruct(dslStruct *trantorpbtypes.ClientProgress) {
	cp.ClientTrackers = make(map[tt.ClientID]*DeliveredTXs, len(dslStruct.Progress))
	for clientID, deliveredTXs := range dslStruct.Progress {
		cp.ClientTrackers[clientID] = DeliveredTXsFromDslStruct(
			deliveredTXs,
			logging.Decorate(cp.logger, "", "clID", clientID),
		)
	}
}

func (cp *ClientProgress) DslStruct() *trantorpbtypes.ClientProgress {
	ds := make(map[tt.ClientID]*trantorpbtypes.DeliveredTXs, len(cp.ClientTrackers))
	for clientID, clientTracker := range cp.ClientTrackers {
		ds[clientID] = clientTracker.DslStruct()
	}

	return &trantorpbtypes.ClientProgress{Progress: ds}
}

func FromPb(pb *trantorpb.ClientProgress, logger logging.Logger) *ClientProgress {
	cp := NewClientProgress(logger)
	cp.LoadPb(pb)
	return cp
}

func FromDslStruct(dslStruct *trantorpbtypes.ClientProgress, logger logging.Logger) *ClientProgress {
	cp := NewClientProgress(logger)
	cp.LoadDslStruct(dslStruct)
	return cp
}
