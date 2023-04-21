package clientprogress

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// DeliveredReqs tracks the watermarks of delivered transactions for a single client.
type DeliveredReqs struct {

	// LowWM is the lowest watermark of the client.
	lowWM tt.ReqNo

	// Set of request numbers indicating whether the corresponding transaction has been delivered.
	// In addition, all request numbers strictly smaller than LowWM are considered delivered.
	delivered map[tt.ReqNo]struct{}

	// Logger to use for logging output.
	logger logging.Logger
}

// EmptyDeliveredReqs allocates and returns a new DeliveredReqs.
func EmptyDeliveredReqs(logger logging.Logger) *DeliveredReqs {
	return &DeliveredReqs{
		lowWM:     0,
		delivered: make(map[tt.ReqNo]struct{}),
		logger:    logger,
	}
}

func DeliveredReqsFromPb(pb *commonpb.DeliveredReqs, logger logging.Logger) *DeliveredReqs {
	dr := EmptyDeliveredReqs(logger)
	dr.lowWM = tt.ReqNo(pb.LowWm)
	for _, reqNo := range pb.Delivered {
		dr.delivered[tt.ReqNo(reqNo)] = struct{}{}
	}
	return dr
}

// Add adds a request number that is considered delivered to the DeliveredReqs.
// Returns true if the request number has been added now (after not being previously present).
// Returns false if the request number has already been added before the call to Add.
func (dr *DeliveredReqs) Add(reqNo tt.ReqNo) bool {

	if reqNo < dr.lowWM {
		dr.logger.Log(logging.LevelDebug, "Request sequence number below client's watermark window.",
			"lowWM", dr.lowWM, "reqNo", reqNo)
		return false
	}

	_, alreadyPresent := dr.delivered[reqNo]
	dr.delivered[reqNo] = struct{}{}
	return !alreadyPresent
}

// GarbageCollect reduces the memory footprint of the DeliveredReqs
// by deleting a contiguous prefix of delivered request numbers
// and increasing the low watermark accordingly.
// Returns the new low watermark.
func (dr *DeliveredReqs) GarbageCollect() tt.ReqNo {

	for _, ok := dr.delivered[dr.lowWM]; ok; _, ok = dr.delivered[dr.lowWM] {
		delete(dr.delivered, dr.lowWM)
		dr.lowWM++
	}

	return dr.lowWM
}

func (dr *DeliveredReqs) Pb() *commonpb.DeliveredReqs {
	delivered := make([]uint64, len(dr.delivered))
	for i, reqNo := range maputil.GetSortedKeys(dr.delivered) {
		delivered[i] = reqNo.Pb()
	}

	return &commonpb.DeliveredReqs{
		LowWm:     dr.lowWM.Pb(),
		Delivered: delivered,
	}
}
