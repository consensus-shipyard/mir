package batchfetcher

import (
	"fmt"

	bfevents "github.com/filecoin-project/mir/pkg/batchfetcher/events"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"

	availabilitypbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	apbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	bfeventstypes "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/events"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"

	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"

	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// NewModule returns a new batch fetcher module.
// The batch fetcher receives events output by the ordering protocol (e.g. ISS)
// and relays them to the application in the same order.
// It replaces the DeliverCert events from the input stream by the corresponding ProvideTransactions
// that it obtains from the availability layer.
// It keeps track of the current epoch (by observing the relayed NewEpoch events)
// and automatically requests the transactions from the correct instance of the availability module.
//
// The batch fetcher also deduplicates the transactions, guaranteeing that each transaction
// is output only the first time it appears in a batch.
// For this purpose, the batch fetcher maintains information about which transactions have been delivered
// and provides it to the checkpoint module when relaying a state snapshot request to the application.
// Analogously, when relaying a RestoreState event, it restores its state (including the delivered transactions)
// using the relayed information.
func NewModule(mc *ModuleConfig, epochNr t.EpochNr, clientProgress *clientprogress.ClientProgress, logger logging.Logger) modules.Module {
	m := dsl.NewModule(mc.Self)
	// Queue of output events. It is required for buffering events being relayed
	// in case a DeliverCert event received earlier has not yet been transformed to a ProvideTransactions event.
	// In such a case, events received later must not be relayed until the pending certificate has been resolved.
	var output outputQueue

	// filterDuplicates takes a NewOrderedBatch event and removes all the contained transactions
	// that have already been added to the clientProgress, i.e., that have already been delivered.
	// filterDuplicates modification performs the modification in-place, on the provided batch.
	// It is applied to each transaction batch immediately before delivering it to the application.
	filterDuplicates := func(newOrderedBatch *batchfetcherpb.NewOrderedBatch) {

		newTxs := make([]*requestpb.Request, 0, len(newOrderedBatch.Txs))

		for _, tx := range newOrderedBatch.Txs {

			// Convenience variables
			clID := t.ClientID(tx.ClientId)
			reqNo := t.ReqNo(tx.ReqNo)

			// Only keep transaction if it has not yet been delivered.
			if clientProgress.Add(clID, reqNo) {
				newTxs = append(newTxs, tx)
			}
		}

		// Replace the original list of transactions by the filtered one.
		newOrderedBatch.Txs = newTxs
	}

	// The NewEpoch handler updates the current epoch number and forwards the event to the output.
	eventpbdsl.UponNewEpoch(m, func(newEpochNr t.EpochNr) error {
		epochNr = newEpochNr
		output.Enqueue(&outputItem{
			event: events.NewEpoch(mc.Destination, epochNr),
		})
		output.Flush(m)
		return nil
	})

	// The DeliverCert handler requests the transactions referenced by the received availability certificate
	// from the availability layer.
	eventpbdsl.UponDeliverCert(m, func(sn t.SeqNr, cert *apbtypes.Cert) error {
		// Create an empty output item and enqueue it immediately.
		// Actual output will be delayed until the transactions have been received.
		// This is necessary to preserve the order of incoming and outgoing events.
		item := outputItem{
			event: nil,

			// At the time of delivering the batch,
			// filter out transactions that have already been delivered in previous batches.
			// Note that this must be done immediately before delivering the batch,
			// NOT on reception of the transaction payloads.
			// (Otherwise, delivering the transaction payloads from the availability module
			// in different order at different nodes would lead to inconsistencies).
			f: func(e *eventpb.Event) {
				// Casting event to the NewOrderedBatch type is safe,
				// because no other event type is ever saved in an output item created at certificate delivery.
				filterDuplicates(e.
					Type.(*eventpb.Event_BatchFetcher).BatchFetcher.
					Type.(*batchfetcherpb.Event_NewOrderedBatch).NewOrderedBatch)
			},
		}
		output.Enqueue(&item)

		//TODO cleanup check for empty certificates and make consistent across modules
		if cert.Type == nil {
			// Skip fetching transactions for padding certificates.
			// Directly deliver an empty batch instead.
			item.event = bfeventstypes.NewOrderedBatch(mc.Destination, []*requestpbtypes.Request{}).Pb()
			output.Flush(m)
		} else {
			// If this is a proper certificate, request transactions from the availability layer.
			availabilitypbdsl.RequestTransactions(
				m,
				mc.Availability.Then(t.ModuleID(fmt.Sprintf("%v", epochNr))),
				cert,
				&txRequestContext{queueItem: &item},
			)
		}

		return nil
	})

	// The AppSnapshotRequest handler triggers a ClientProgress event (for the checkpointing protocol)
	// and forwards the original snapshot request event to the output.
	eventpbdsl.UponAppSnapshotRequest(m, func(replyTo t.ModuleID) error {
		// Save the number of the epoch when the AppSnapshotRequest has been received.
		// This is necessary in case the epoch number changes
		// by the time the AppSnapshotRequest event is output and the hook function (added below) executed.
		// Forward the original event to the output.
		output.Enqueue(&outputItem{
			event: events.AppSnapshotRequest(mc.Destination, replyTo),

			// At the time of forwarding, submit the client progress to the checkpointing protocol.
			f: func(_ *eventpb.Event) {
				clientProgress.GarbageCollect()
				//commonpbdsl.ClientProgress(m, mc.Checkpoint, clientProgress.ClientTrackers)
				dsl.EmitEvent(m, bfevents.ClientProgress(
					mc.Checkpoint.Then(t.ModuleID(fmt.Sprintf("%v", epochNr))),
					clientProgress.Pb(),
				))
			},
		})
		output.Flush(m)

		return nil
	})

	// The AppRestoreState handler restores the batch fetcher's state from a checkpoint
	// and forwards the event to the application, so it can restore its state too.
	eventpbdsl.UponAppRestoreState(m, func(mirChkp *checkpointpbtypes.StableCheckpoint) error {

		chkp := checkpoint.StableCheckpointFromPb(mirChkp.Pb())

		// Update current epoch number.
		epochNr = chkp.Epoch()

		// Load client progress.
		clientProgress = chkp.ClientProgress(logger)

		// Reset output event queue.
		// This is necessary to prune any pending output to the application
		// that pertains to the epochs before this checkpoint.
		output = outputQueue{}

		// Forward the RestoreState event to the application.
		// We can output it directly without passing through the queue,
		// since we've just reset it and know this would be its first and only item.
		eventpbdsl.AppRestoreState(m, mc.Destination, mirChkp)

		return nil
	})

	// The ProvideTransactions handler filters the received transaction batch,
	// removing all transactions that have been previously delivered,
	// assigns the remaining transactions to the corresponding output item
	// (the one created on reception of the corresponding availability certificate in DeliverCert)
	// and flushes the output stream.
	availabilitypbdsl.UponProvideTransactions(m, func(txs []*requestpbtypes.Request, context *txRequestContext) error {

		// Note that not necessarily all transactions will be part of the final batch.
		// When the event leaves the output buffer, duplicates will be filtered out.
		context.queueItem.event = bfeventstypes.NewOrderedBatch(mc.Destination, txs).Pb()
		output.Flush(m)
		return nil
	})

	// Explicitly ignore Init event. This prevents forwarding it to the destination module.
	eventpbdsl.UponInit(m, func() error {
		return nil
	})

	// All other events simply pass through the batch fetcher unchanged (except their destination module).
	dsl.UponOtherEvent(m, func(ev *eventpb.Event) error {
		output.Enqueue(&outputItem{
			event: events.Redirect(ev, mc.Destination),
		})
		output.Flush(m)
		return nil
	})

	return m
}

// txRequestContext saves the context of requesting transactions from the availability layer.
type txRequestContext struct {
	queueItem *outputItem
}
