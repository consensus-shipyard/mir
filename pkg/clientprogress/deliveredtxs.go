package clientprogress

import (
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/trantorpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// DeliveredTXs tracks the watermarks of delivered transactions for a single client.
type DeliveredTXs struct {

	// LowWM is the lowest watermark of the client.
	lowWM tt.TxNo

	// Set of transaction numbers indicating whether the corresponding transaction has been delivered.
	// In addition, all transaction numbers strictly smaller than LowWM are considered delivered.
	delivered map[tt.TxNo]struct{}

	// Logger to use for logging output.
	logger logging.Logger
}

// EmptyDeliveredTXs allocates and returns a new DeliveredTXs.
func EmptyDeliveredTXs(logger logging.Logger) *DeliveredTXs {
	return newDeliveredTXs(logger, 0)
}

func newDeliveredTXs(logger logging.Logger, capacity int) *DeliveredTXs {
	return &DeliveredTXs{
		lowWM:     0,
		delivered: make(map[tt.TxNo]struct{}, capacity),
		logger:    logger,
	}
}

func DeliveredTXsFromPb(pb *trantorpb.DeliveredTXs, logger logging.Logger) *DeliveredTXs {
	dr := newDeliveredTXs(logger, len(pb.Delivered))
	dr.lowWM = tt.TxNo(pb.LowWm)
	for _, txNo := range pb.Delivered {
		dr.delivered[tt.TxNo(txNo)] = struct{}{}
	}
	return dr
}

func DeliveredTXsFromDslStruct(dslStruct *trantorpbtypes.DeliveredTXs, logger logging.Logger) *DeliveredTXs {
	dr := newDeliveredTXs(logger, len(dslStruct.Delivered))
	dr.lowWM = dslStruct.LowWm
	for _, txNo := range dslStruct.Delivered {
		dr.delivered[txNo] = struct{}{}
	}
	return dr
}

// Contains returns true if the given txNo has already been added.
func (dt *DeliveredTXs) Contains(txNo tt.TxNo) bool {

	if txNo < dt.lowWM {
		return true
	}

	_, alreadyPresent := dt.delivered[txNo]
	return alreadyPresent
}

// Add adds a transaction number that is considered delivered to the DeliveredTXs.
// Returns true if the transaction number has been added now (after not being previously present).
// Returns false if the transaction number has already been added before the call to Add.
func (dt *DeliveredTXs) Add(txNo tt.TxNo) bool {

	if txNo < dt.lowWM {
		dt.logger.Log(logging.LevelDebug, "Transaction number below client's watermark window.",
			"lowWM", dt.lowWM, "txNo", txNo)
		return false
	}

	_, alreadyPresent := dt.delivered[txNo]
	dt.delivered[txNo] = struct{}{}
	return !alreadyPresent
}

// GarbageCollect reduces the memory footprint of the DeliveredTXs
// by deleting a contiguous prefix of delivered transaction numbers
// and increasing the low watermark accordingly.
// Returns the new low watermark.
func (dt *DeliveredTXs) GarbageCollect() tt.TxNo {

	for _, ok := dt.delivered[dt.lowWM]; ok; _, ok = dt.delivered[dt.lowWM] {
		delete(dt.delivered, dt.lowWM)
		dt.lowWM++
	}

	return dt.lowWM
}

func (dt *DeliveredTXs) Pb() *trantorpb.DeliveredTXs {
	delivered := make([]uint64, len(dt.delivered))
	for i, txNo := range maputil.GetSortedKeys(dt.delivered) {
		delivered[i] = txNo.Pb()
	}

	return &trantorpb.DeliveredTXs{
		LowWm:     dt.lowWM.Pb(),
		Delivered: delivered,
	}
}

func (dt *DeliveredTXs) DslStruct() *trantorpbtypes.DeliveredTXs {
	delivered := make([]tt.TxNo, len(dt.delivered))
	for i, txNo := range maputil.GetSortedKeys(dt.delivered) {
		delivered[i] = txNo
	}

	return &trantorpbtypes.DeliveredTXs{
		LowWm:     dt.lowWM,
		Delivered: delivered,
	}
}
