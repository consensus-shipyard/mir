// The batch creation implemented here works as follows:
//
// The mempool keeps the incoming transactions in a list in the order of arrival.
// At the same time, the mempool keeps track of all the transactions that already have been delivered.
// This information is stored in the ClientProgress object that is updated at the beginning of every epoch.
// The mempool never keeps transactions that have already been delivered
// and prunes the stored ones each time ClientProgress is updated.
//
// In order to not emit a transaction twice in the same epoch, the mempool keeps track of epochs
// and uses iterators to retrieve transactions from the stored transaction list.
// In each epoch, after updating the ClientProgress and pruning delivered transactions,
// it creates a new iterator that it uses for reading the remaining stored transactions from the beginning.
//
// If another module that has already advanced to a new epoch requests a batch
// while the mempool still has not advanced to that epoch, it risks using an old iterator for a new batch request.
// To this end, batch requests are also tagged with an epoch and the mempool only handles them in their proper epoch,
// buffers them if necessary.

package formbatchesint

import (
	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/common"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	mppbdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mppbevents "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	mppbtypes "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	timertypes "github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/indexedlist"
)

type State struct {
	*common.State

	// The current epoch.
	// If a batch requests from a higher epoch is received, it needs to be buffered until its epoch is reached.
	Epoch tt.EpochNr

	// Progress made by all clients so far.
	// This data structure is used to avoid storing transactions that have already been delivered.
	ClientProgress *clientprogress.ClientProgress

	// Combined total payload size of all the transactions in the mempool.
	TotalPayloadSize int

	// The cummulative number of payload bytes of transactions that have not been output in batches in this epoch.
	// At the beginning of each epoch, this value is reset to TotalPayloadSize.
	UnproposedPayloadSize int

	// Number of transactions that have not yet been output in batches in the current epoch.
	// At the beginning of each epoch, this value is reset to Transactions.Len().
	NumUnproposed int

	// Iterator over the list transactions in the mempool.
	// At the start of each epoch, the iterator is reset to the start of the list.
	// This is necessary for re-emitting transactions that already have been emitted in the previous epoch
	// But failed to be agreed upon.
	Iterator *indexedlist.Iterator[tt.TxID, *trantorpbtypes.Transaction]

	// EarlyBatchRequests stores batch requests with a higher epoch number than the current epoch.
	// In Trantor, this can happen in a corner case
	// when advancing to a new epoch is delayed by the processing of the batch fetcher.
	// Note that these are different from PendingBatchRequests,
	// which are this epoch's batch requests waiting for a batch to fill (or a timeout)
	EarlyBatchRequests map[tt.EpochNr][]*mppbtypes.RequestBatchOrigin

	// Pending batch requests, i.e., this epoch's batch requests that have not yet been satisfied.
	// They are indexed by order of arrival.
	// Note that this only concerns batch requests from the current epoch.
	// Requests from future epochs are buffered separately.
	PendingBatchRequests map[int]*mppbtypes.RequestBatchOrigin

	// Index of the oldest pending batch request.
	FirstPendingBatchReqID int

	// Index to assign to the next new pending batch request.
	NextPendingBatchReqID int
}

// IncludeBatchCreation registers event handlers for processing NewRequests and RequestBatch events.
func IncludeBatchCreation( // nolint:gocognit
	m dsl.Module,
	mc common.ModuleConfig,
	params *common.ModuleParams,
	commonState *common.State,
	logger logging.Logger,
) {
	state := &State{
		State:                  commonState,
		Epoch:                  0,
		ClientProgress:         clientprogress.NewClientProgress(logger),
		TotalPayloadSize:       0,
		UnproposedPayloadSize:  0,
		NumUnproposed:          0,
		Iterator:               commonState.Transactions.Iterator(0),
		EarlyBatchRequests:     make(map[tt.EpochNr][]*mppbtypes.RequestBatchOrigin),
		PendingBatchRequests:   make(map[int]*mppbtypes.RequestBatchOrigin),
		FirstPendingBatchReqID: 0,
		NextPendingBatchReqID:  0,
	}

	// cutBatch creates a new batch from whatever transactions are available in the mempool
	// (even if the batch ends up empty), and emits it as an event associated with the given origin.
	cutBatch := func(origin *mppbtypes.RequestBatchOrigin) {

		batchSize := 0
		txCount := 0

		txIDs, txs, _ := state.Iterator.NextWhile(func(_ tt.TxID, tx *trantorpbtypes.Transaction) bool {
			if txCount < params.MaxTransactionsInBatch && batchSize+len(tx.Data) <= params.MaxPayloadInBatch {
				txCount++
				state.NumUnproposed--
				batchSize += len(tx.Data)
				state.UnproposedPayloadSize -= len(tx.Data)
				return true
			}
			return false
		})

		// Note that a batch may be empty.
		mppbdsl.NewBatch(m, origin.Module, txIDs, txs, origin)
	}

	// storePendingRequest creates an entry for a new batch request in the pending batch request list.
	// It generates a unique ID for the pending request that can be used to associate a timeout with it.
	storePendingRequest := func(origin *mppbtypes.RequestBatchOrigin) int {
		state.PendingBatchRequests[state.NextPendingBatchReqID] = origin
		state.NextPendingBatchReqID++
		return state.NextPendingBatchReqID - 1
	}

	// Returns true if the mempool contains enough transactions for a full batch.
	haveFullBatch := func() bool {
		return state.NumUnproposed >= params.MaxTransactionsInBatch ||
			state.UnproposedPayloadSize >= params.MaxPayloadInBatch
	}

	// Cuts a new batch for a batch request with the given ID and updates the corresponding internal data structures.
	servePendingReq := func(batchReqID int) {

		// Serve the given pending batch request and remove it from the pending list.
		cutBatch(state.PendingBatchRequests[batchReqID])
		delete(state.PendingBatchRequests, batchReqID)

		// Advance the pointer to the first pending batch request.
		for _, ok := state.PendingBatchRequests[state.FirstPendingBatchReqID]; !ok && state.FirstPendingBatchReqID < state.NextPendingBatchReqID; _, ok = state.PendingBatchRequests[state.FirstPendingBatchReqID] {
			state.FirstPendingBatchReqID++
		}
	}

	handleBatchRequest := func(origin *mppbtypes.RequestBatchOrigin) {
		if haveFullBatch() {
			cutBatch(origin)
		} else {
			reqID := storePendingRequest(origin)
			eventpbdsl.TimerDelay(m,
				mc.Timer,
				[]*eventpbtypes.Event{mppbevents.BatchTimeout(mc.Self, uint64(reqID))},
				timertypes.Duration(params.BatchTimeout),
			)
		}
	}

	mppbdsl.UponNewEpoch(m,
		func(epochNr tt.EpochNr, clientProgress *trantorpbtypes.ClientProgress) error {

			// Update the local view of the epoch number.
			state.Epoch = epochNr

			// Garbage-collect the old iterator and create a new one.
			state.Transactions.GarbageCollect(tt.RetentionIndex(epochNr))
			state.Iterator = state.Transactions.Iterator(tt.RetentionIndex(epochNr))

			// Update client progress and prune delivered transactions.
			// TODO: This might be inefficient, especially if there are many transactions in the mempool.
			//   A potential solution would be to keep an index of pending transactions similar to
			//   ClientProgress - for each client, list of pending transactions sorted by TxNo - that
			//   would make pruning significantly more efficient.
			state.ClientProgress.LoadPb(clientProgress.Pb())
			_, removedTXs := state.Transactions.RemoveSelected(func(_ tt.TxID, tx *trantorpbtypes.Transaction) bool {
				return state.ClientProgress.Contains(tx.ClientId, tx.TxNo)
			})
			for _, tx := range removedTXs {
				state.TotalPayloadSize -= len(tx.Data)
			}

			// Reset trackers of unproposed transactions.
			state.NumUnproposed = state.Transactions.Len()
			state.UnproposedPayloadSize = state.TotalPayloadSize

			// Garbage-collect outdated buffered early batch requests, if any,
			// and process the buffered up-to-date ones.
			for epoch, batchReqs := range state.EarlyBatchRequests {
				if epoch < state.Epoch {
					delete(state.EarlyBatchRequests, epoch)
				} else if epoch == state.Epoch {
					for _, batchReq := range batchReqs {
						handleBatchRequest(batchReq)
					}
				}
			}

			return nil
		},
	)

	mppbdsl.UponNewTransactions(m, func(txsReceived []*trantorpbtypes.Transaction) error {

		// Filter out transactions that have already been delivered to the application.
		txsToSave := make([]*trantorpbtypes.Transaction, 0, len(txsReceived))
		for _, tx := range txsReceived {

			// Only save transactions with payload not larger than the batch limit
			// (as they would not fit in any batch, even if no other transactions were present).
			if len(tx.Data) <= params.MaxPayloadInBatch {
				txsToSave = append(txsToSave, tx)
			} else {
				logger.Log(logging.LevelWarn, "Discarding transaction. Payload larger than batch limit.",
					"MaxPayloadInBatch", params.MaxPayloadInBatch, "PayloadSize", len(tx.Data))
			}

		}
		// TODO: Introduce a client watermark window size and also ignore transactions
		//   that are past the watermark window.

		// Compute IDs of the new transactions
		if len(txsToSave) > 0 {
			mppbdsl.RequestTransactionIDs(m, mc.Self, txsToSave, &requestTxIDsContext{txsToSave})
		}

		return nil
	})

	mppbdsl.UponTransactionIDsResponse(m, func(txIDs []tt.TxID, context *requestTxIDsContext) error {

		_, addedTxs := state.Transactions.Append(txIDs, context.txs)
		for _, tx := range addedTxs {

			// Discard transactions that have already been delivered in a previous epoch.
			if state.ClientProgress.Contains(tx.ClientId, tx.TxNo) {
				continue
			}

			state.TotalPayloadSize += len(tx.Data)
			state.UnproposedPayloadSize += len(tx.Data)
			state.NumUnproposed++
		}

		for haveFullBatch() && state.FirstPendingBatchReqID < state.NextPendingBatchReqID {
			servePendingReq(state.FirstPendingBatchReqID)
		}

		return nil
	})

	mppbdsl.UponRequestBatch(m, func(epoch tt.EpochNr, origin *mppbtypes.RequestBatchOrigin) error {
		if epoch == state.Epoch {
			// Only handle batch requests from the current epoch.
			handleBatchRequest(origin)
		} else if epoch > state.Epoch {
			// Buffer requests from future epochs.
			state.EarlyBatchRequests[epoch] = append(state.EarlyBatchRequests[epoch], origin)
			// TODO: Write tests that explore this code path.
		} // (Requests from past epochs are ignored.)
		return nil
	})

	mppbdsl.UponBatchTimeout(m, func(batchReqID uint64) error {

		reqID := int(batchReqID)

		// Load the request origin.
		_, ok := state.PendingBatchRequests[reqID]

		if ok {
			// If request is still pending, respond to it.
			servePendingReq(reqID)
		} else {
			// Ignore timeout if request has already been served.
			logger.Log(logging.LevelDebug, "Ignoring outdated batch timeout.",
				"batchReqID", reqID)
		}
		return nil
	})
}

// Context data structures

type requestTxIDsContext struct {
	txs []*trantorpbtypes.Transaction
}
