package formbatchesint

import (
	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/common"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	mpdsl "github.com/filecoin-project/mir/pkg/pb/mempoolpb/dsl"
	mpevents "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	mppbtypes "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	timertypes "github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/indexedlist"
)

type State struct {
	*common.State

	// Transactions received, but not yet emitted in any batch.
	NewTxIDs []tt.TxID

	// Combined total payload size of all the transactions in the mempool.
	TotalPayloadSize int

	// Pending batch requests, i.e., batch requests that have not yet been satisfied.
	// They are indexed by order of arrival.
	PendingBatchRequests map[int]*mppbtypes.RequestBatchOrigin

	// Index of the oldest pending batch request.
	FirstPendingBatchReqID int

	// Index to assign to the next new pending batch request.
	NextPendingBatchReqID int
}

// IncludeBatchCreation registers event handlers for processing NewRequests and RequestBatch events.
func IncludeBatchCreation(
	m dsl.Module,
	mc common.ModuleConfig,
	params *common.ModuleParams,
	commonState *common.State,
	logger logging.Logger,
) {
	state := &State{
		State:                  commonState,
		NewTxIDs:               nil,
		TotalPayloadSize:       0,
		PendingBatchRequests:   make(map[int]*mppbtypes.RequestBatchOrigin),
		FirstPendingBatchReqID: 0,
		NextPendingBatchReqID:  0,
	}

	// cutBatch creates a new batch from whatever transactions are available in the mempool
	// (even if the batch ends up empty), and emits it as an event associated with the given origin.
	cutBatch := func(origin *mppbtypes.RequestBatchOrigin) {
		var txIDs []tt.TxID
		var txs []*trantorpbtypes.Transaction
		batchSize := 0

		txCount := 0
		for _, txID := range state.NewTxIDs {
			tx := state.TxByID[txID]

			// Stop adding TXs if count or size limit has been reached.
			if txCount == params.MaxTransactionsInBatch || batchSize+len(tx.Data) > params.MaxPayloadInBatch {
				break
			}

			txIDs = append(txIDs, txID)
			txs = append(txs, tx)
			batchSize += len(tx.Data)
			txCount++
		}

		for _, txID := range state.NewTxIDs[:txCount] {
			delete(state.TxByID, txID)
		}

		// Update the list of pending transactions and their total size.
		state.NewTxIDs = state.NewTxIDs[txCount:]
		state.TotalPayloadSize -= batchSize

		// Note that a batch may be empty.
		mpdsl.NewBatch(m, origin.Module, txIDs, txs, origin)
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
		return len(state.NewTxIDs) >= params.MaxTransactionsInBatch ||
			state.TotalPayloadSize >= params.MaxPayloadInBatch
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

	mpdsl.UponNewTransactions(m, func(txs []*trantorpbtypes.Transaction) error {
		mpdsl.RequestTransactionIDs(m, mc.Self, txs, &requestTxIDsContext{txs})
		return nil
	})

	mpdsl.UponTransactionIDsResponse(m, func(txIDs []tt.TxID, context *requestTxIDsContext) error {
		for i, txID := range txIDs {
			tx := context.txs[i]

			// Discard transactions with payload larger than batch limit
			// (as they would not fit in any batch, even if no other transactions were present).
			if len(tx.Data) > params.MaxPayloadInBatch {
				logger.Log(logging.LevelWarn, "Discarding transaction. Payload larger than batch limit.",
					"MaxPayloadInBatch", params.MaxPayloadInBatch, "PayloadSize", len(tx.Data))
				continue
			}

			state.TxByID[txIDs[i]] = tx
			state.TotalPayloadSize += len(tx.Data)
			state.NewTxIDs = append(state.NewTxIDs, txID)
		}

		for haveFullBatch() && state.FirstPendingBatchReqID < state.NextPendingBatchReqID {
			servePendingReq(state.FirstPendingBatchReqID)
		}

		return nil
	})

	mpdsl.UponRequestBatch(m, func(origin *mppbtypes.RequestBatchOrigin) error {
		if haveFullBatch() {
			cutBatch(origin)
		} else {
			reqID := storePendingRequest(origin)
			eventpbdsl.TimerDelay(m,
				mc.Timer,
				[]*eventpbtypes.Event{mpevents.BatchTimeout(mc.Self, uint64(reqID))},
				timertypes.Duration(params.BatchTimeout),
			)
		}
		return nil
	})

	mpdsl.UponBatchTimeout(m, func(batchReqID uint64) error {

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
