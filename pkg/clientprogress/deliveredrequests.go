package clientprogress

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/trantorpb"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// DeliveredTXs tracks the watermarks of delivered transactions for a single client.
type DeliveredTXs struct {

	// LowWM is the lowest watermark of the client.
	lowWM tt.ReqNo

	// Set of transaction numbers indicating whether the corresponding transaction has been delivered.
	// In addition, all transaction numbers strictly smaller than LowWM are considered delivered.
	delivered map[tt.ReqNo]struct{}

	// Logger to use for logging output.
	logger logging.Logger
}

// EmptyDeliveredTXs allocates and returns a new DeliveredTXs.
func EmptyDeliveredTXs(logger logging.Logger) *DeliveredTXs {
	return &DeliveredTXs{
		lowWM:     0,
		delivered: make(map[tt.ReqNo]struct{}),
		logger:    logger,
	}
}

func DeliveredTXsFromPb(pb *trantorpb.DeliveredTXs, logger logging.Logger) *DeliveredTXs {
	dr := EmptyDeliveredTXs(logger)
	dr.lowWM = tt.ReqNo(pb.LowWm)
	for _, reqNo := range pb.Delivered {
		dr.delivered[tt.ReqNo(reqNo)] = struct{}{}
	}
	return dr
}

// Add adds a transaction number that is considered delivered to the DeliveredTXs.
// Returns true if the transaction number has been added now (after not being previously present).
// Returns false if the transaction number has already been added before the call to Add.
func (dr *DeliveredTXs) Add(reqNo tt.ReqNo) bool {

	if reqNo < dr.lowWM {
		dr.logger.Log(logging.LevelDebug, "Transaction number below client's watermark window.",
			"lowWM", dr.lowWM, "reqNo", reqNo)
		return false
	}

	_, alreadyPresent := dr.delivered[reqNo]
	dr.delivered[reqNo] = struct{}{}
	return !alreadyPresent
}

// GarbageCollect reduces the memory footprint of the DeliveredTXs
// by deleting a contiguous prefix of delivered transaction numbers
// and increasing the low watermark accordingly.
// Returns the new low watermark.
func (dr *DeliveredTXs) GarbageCollect() tt.ReqNo {

	for _, ok := dr.delivered[dr.lowWM]; ok; _, ok = dr.delivered[dr.lowWM] {
		delete(dr.delivered, dr.lowWM)
		dr.lowWM++
	}

	return dr.lowWM
}

func (dr *DeliveredTXs) Pb() *trantorpb.DeliveredTXs {
	delivered := make([]uint64, len(dr.delivered))
	for i, reqNo := range maputil.GetSortedKeys(dr.delivered) {
		delivered[i] = reqNo.Pb()
	}

	return &trantorpb.DeliveredTXs{
		LowWm:     dr.lowWM.Pb(),
		Delivered: delivered,
	}
}
