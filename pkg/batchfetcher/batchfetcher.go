package batchfetcher

import (
	"fmt"

	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"

	bfevents "github.com/filecoin-project/mir/pkg/batchfetcher/events"
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
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
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
func NewModule(mc *ModuleConfig, epochNr t.EpochNr, clientProgress *clientprogress.ClientProgress) modules.Module {
	m := dsl.NewModule(mc.Self)

	// Queue of output events. It is required for buffering events being relayed
	// in case a DeliverCert event received earlier has not yet been transformed to a ProvideTransactions event.
	// In such a case, events received later must not be relayed until the pending certificate has been resolved.
	var output outputQueue

	// The NewEpoch handler updates the current epoch number and forwards the event to the output.
	eventpbdsl.UponNewEpoch(m, func(epochNr t.EpochNr) error {
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
		item := outputItem{event: nil}
		output.Enqueue(&item)

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
	eventpbdsl.UponAppRestoreState(m, func(chkp *checkpointpbtypes.StableCheckpoint) error {
		// Update current epoch number.
		epochNr = t.EpochNr(chkp.Snapshot.EpochData.EpochConfig.EpochNr)

		// Load client progress.
		clientProgress.LoadPb(chkp.Snapshot.EpochData.ClientProgress)

		// Reset output event queue.
		// This is necessary to prune any pending output to the application
		// that pertains to the epochs before this checkpoint.
		output = outputQueue{}

		// Forward the RestoreState event to the application.
		// We can output it directly without passing through the queue,
		// since we've just reset it and know this would be its first and only item.
		eventpbdsl.AppRestoreState(m, mc.Destination, chkp)

		return nil
	})

	// The ProvideTransactions handler filters the received transaction batch,
	// removing all transactions that have been previously delivered,
	// assigns the remaining transactions to the corresponding output item
	// (the one created on reception of the corresponding availability certificate in DeliverCert)
	// and flushes the output stream.
	availabilitypbdsl.UponProvideTransactions(m, func(txs []*requestpbtypes.Request, context *txRequestContext) error {

		// Filter out transactions that already have been delivered
		newTxs := make([]*requestpb.Request, 0, len(txs))
		for _, req := range txs {
			// Runs for each received transaction.

			// Convenience variables
			clID := req.ClientId
			reqNo := req.ReqNo

			// Only keep request if it has not yet been delivered.
			if clientProgress.Add(clID, reqNo) {
				newTxs = append(newTxs, req.Pb())
			}
		}

		context.queueItem.event = bfevents.NewOrderedBatch(mc.Destination, newTxs)
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
