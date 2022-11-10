package smr

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/reactivex/rxgo/v2"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// TODO: this reads the whole event trace into memory. Consider directly creating and returning an Obsrbavle
// that would read the events on demand.
func eventTrace(eventLogFile string) (events []*eventpb.Event, retErr error) {

	// Open event log file.
	f, err := os.Open(eventLogFile)
	if err != nil {
		return nil, err
	}

	// Schedule closing file.
	defer func() {
		err := f.Close()
		if retErr == nil {
			retErr = err
		}
	}()

	reader, err := eventlog.NewReader(f)
	if err != nil {
		return nil, err
	}

	trace, err := reader.ReadAllEvents()
	if err != nil {
		return nil, err
	}

	return trace, nil
}

func checkEventTraces(eventLogFiles map[t.NodeID]string, requestsSubmitted int) error {
	// Create Observables for event traces, one per node.
	// The event traces are restricted (filtered) to include only those events
	// that deliver transaction batches to the application.
	deliveredBatches := make(map[t.NodeID]rxgo.Observable, len(eventLogFiles))
	for nodeID, eventLogFile := range eventLogFiles {
		trace, err := eventTrace(eventLogFile)
		if err != nil {
			return fmt.Errorf("failed to load event trace from file %s: %w", eventLogFile, err)
		}
		deliveredBatches[nodeID] = rxgo.Just(trace)().Filter(isNewOrderedBatch).Map(extractBatch)
	}

	// Check that all nodes delivered the same sequence of batches as the first one.
	// To do that, compare all nodes' sequences of delivered batches to the one of the first node.
	firstNodeID := maputil.GetSortedKeys(deliveredBatches)[0]
	firstTrace := deliveredBatches[firstNodeID]
	for nodeID, db := range deliveredBatches {

		// Track the number of the batch (only for reporting errors).
		i := 0

		// To compare two traces, we use the "Zip" operator that combines, at each position,
		// the two elements (batches in our case) in the respective trace.
		// The combination function produces an error if the two batches are not equal.
		err := firstTrace.ZipFromIterable(
			db,
			func(_ context.Context, b1 interface{}, b2 interface{}) (interface{}, error) {
				if err := batchesEqual(b1, b2); err != nil {
					return nil, fmt.Errorf("batches not equal at index %d for nodes %v and %v: %w",
						i,
						firstNodeID,
						nodeID,
						err,
					)
				}
				return b1, nil
			},
		).Error()
		if err != nil {
			return fmt.Errorf("delivered batch history mismatch: %w", err)
		}
	}

	// The code above only checks whether the prefix of all delivered batch sequences
	// matches the whole sequence of batches the first node delivered (this is how the "Zip" operator works).
	// We still need to check whether any other nodes did not deliver additional batches.

	// Count the number of batches delivered by the first node.
	count, err := firstTrace.Count().Get()
	if err != nil {
		return fmt.Errorf("could not get number of delivered batches: %w", err)
	}
	if count.Error() {
		return fmt.Errorf("could not read number of delivered batches: %w", count.E)
	}

	// Check if all nodes delivered the same number of batches (as the first node).
	for _, db := range deliveredBatches {
		cnt, err := db.Count().Get()
		if err != nil {
			return fmt.Errorf("could not get number of delivered batches: %w", err)
		}
		if cnt.Error() {
			return fmt.Errorf("could not read number of delivered batches: %w", cnt.E)
		}
		if count.V.(int64) != cnt.V.(int64) {
			return fmt.Errorf("delivered different number of batches: %v and %v", count.V, cnt.V)
		}
	}

	// Check if each node delivered as many requests as were submitted.
	for _, db := range deliveredBatches {
		cnt, err := db.Map(func(_ context.Context, b interface{}) (interface{}, error) {
			return int64(len(b.(*batchfetcherpb.NewOrderedBatch).Txs)), nil
		}).SumInt64().Get()
		if err != nil {
			return fmt.Errorf("could not get number of delivered transactions: %w", err)
		}
		if cnt.Error() {
			return fmt.Errorf("could not read number of delivered batches: %w", cnt.E)
		}
		if cnt.V.(int64) != int64(requestsSubmitted) {
			return fmt.Errorf("different number of submitted and delivered transactions: %v and %v",
				int64(requestsSubmitted), cnt.V)
		}
	}

	return nil
}

func isNewOrderedBatch(i interface{}) bool {
	bfEvent, ok := i.(*eventpb.Event).Type.(*eventpb.Event_BatchFetcher)
	if !ok {
		return false
	}
	_, ok = bfEvent.BatchFetcher.Type.(*batchfetcherpb.Event_NewOrderedBatch)
	return ok
}

// Does not check any more whether item is a NewOrderedBatch event.
func extractBatch(_ context.Context, i interface{}) (interface{}, error) {
	return i.(*eventpb.Event).Type.(*eventpb.Event_BatchFetcher).
		BatchFetcher.Type.(*batchfetcherpb.Event_NewOrderedBatch).
		NewOrderedBatch, nil
}

func batchesEqual(i1 interface{}, i2 interface{}) error {
	b1 := i1.(*batchfetcherpb.NewOrderedBatch)
	b2 := i2.(*batchfetcherpb.NewOrderedBatch)

	// Check that the bathces contain the same number of transactions.
	if len(b1.Txs) != len(b2.Txs) {
		return fmt.Errorf("batches have different lengths: %d and %d", len(b1.Txs), len(b2.Txs))
	}

	for i, tx := range b1.Txs {
		if err := txEqual(tx, b2.Txs[i]); err != nil {
			return err
		}
	}

	return nil
}

func txEqual(tx1 *requestpb.Request, tx2 *requestpb.Request) error {
	if tx1.ClientId != tx2.ClientId {
		return fmt.Errorf("client ID mismatch: %v and %v", tx1.ClientId, tx2.ClientId)
	}
	if tx1.ReqNo != tx2.ReqNo {
		return fmt.Errorf("request number mismatch: %v and %v", tx1.ReqNo, tx2.ReqNo)
	}
	if tx1.Type != tx2.Type {
		return fmt.Errorf("type mismatch: %v and %v", tx1.Type, tx2.Type)
	}
	if !bytes.Equal(tx1.Data, tx2.Data) {
		return fmt.Errorf("payload mismatch: %v and %v", tx1.Data, tx2.Data)
	}

	return nil
}
