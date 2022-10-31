package clientprogress

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// DeliveredReqs tracks the watermarks of delivered transactions for a single client.
type DeliveredReqs struct {

	// LowWM is the lowest watermark of the client.
	lowWM t.ReqNo

	// Set of request numbers indicating whether the corresponding transaction has been delivered.
	// In addition, all request numbers strictly smaller than LowWM are considered delivered.
	delivered map[t.ReqNo]struct{}

	// Logger to use for logging output.
	logger logging.Logger
}

// EmptyDeliveredReqs allocates and returns a new DeliveredReqs.
func EmptyDeliveredReqs(logger logging.Logger) *DeliveredReqs {
	return &DeliveredReqs{
		lowWM:     0,
		delivered: make(map[t.ReqNo]struct{}),
		logger:    logger,
	}
}

func DeliveredReqsFromPb(pb *commonpb.DeliveredReqs, logger logging.Logger) *DeliveredReqs {
	dr := EmptyDeliveredReqs(logger)
	dr.lowWM = t.ReqNo(pb.LowWm)
	for _, reqNo := range pb.Delivered {
		dr.delivered[t.ReqNo(reqNo)] = struct{}{}
	}
	return dr
}

// Add adds a request number that is considered delivered to the DeliveredReqs.
// Returns true if the request number has been added now (after not being previously present).
// Returns false if the request number has already been added before the call to Add.
func (cwmt *DeliveredReqs) Add(reqNo t.ReqNo) bool {

	if reqNo < cwmt.lowWM {
		cwmt.logger.Log(logging.LevelDebug, "Request sequence number below client's watermark window.",
			"lowWM", cwmt.lowWM, "clId", cwmt.lowWM, "reqNo", reqNo)
		return false
	}

	_, alreadyPresent := cwmt.delivered[reqNo]
	cwmt.delivered[reqNo] = struct{}{}
	return !alreadyPresent
}

// GarbageCollect reduces the memory footprint of the DeliveredReqs
// by deleting a contiguous prefix of delivered request numbers
// and increasing the low watermark accordingly.
// Returns the new low watermark.
func (cwmt *DeliveredReqs) GarbageCollect() t.ReqNo {

	for _, ok := cwmt.delivered[cwmt.lowWM]; ok; _, ok = cwmt.delivered[cwmt.lowWM] {
		delete(cwmt.delivered, cwmt.lowWM)
		cwmt.lowWM++
	}

	return cwmt.lowWM
}

func (cwmt *DeliveredReqs) Pb() *commonpb.DeliveredReqs {
	delivered := make([]uint64, len(cwmt.delivered))
	for i, reqNo := range maputil.GetSortedKeys(cwmt.delivered) {
		delivered[i] = reqNo.Pb()
	}

	return &commonpb.DeliveredReqs{
		LowWm:     cwmt.lowWM.Pb(),
		Delivered: delivered,
	}
}
