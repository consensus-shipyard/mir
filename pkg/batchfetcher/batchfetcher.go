package batchfetcher

import (
	"fmt"

	availabilitydsl "github.com/filecoin-project/mir/pkg/availability/dsl"
	bfevents "github.com/filecoin-project/mir/pkg/batchfetcher/events"
	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
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
// TODO: Implement proper state initialization and transfer,
// comprising not just the epoch number, but also the client watermarks.
func NewModule(mc *ModuleConfig, epochNr t.EpochNr, clientProgress *clientprogress.ClientProgress) modules.Module {
	m := dsl.NewModule(mc.Self)

	// Queue of output events. It is required for buffering events being relayed
	// in case a DeliverCert event received earlier has not yet been transformed to a ProvideTransactions event.
	// In such a case, events received later must not be relayed until the pending certificate has been resolved.
	var output outputQueue

	// The NewEpoch handler updates the current epoch number and forwards the event to the output.
	dsl.UponEvent[*eventpb.Event_NewEpoch](m, func(newEpoch *eventpb.NewEpoch) error {
		epochNr = t.EpochNr(newEpoch.EpochNr)
		output.Enqueue(&outputItem{
			event: events.NewEpoch(mc.Destination, t.EpochNr(newEpoch.EpochNr)),
		})
		output.Flush(m)
		return nil
	})

	// The ClientProgress handler restores the module's view of the client progress.
	// This happens when state is being loaded from a checkpoint.
	dsl.UponEvent[*eventpb.Event_BatchFetcher](m, func(bfEvent *batchfetcherpb.Event) error {
		// TODO: Write a batchfetcher DSL package and move the boilerplate there.
		switch e := bfEvent.Type.(type) {
		case *batchfetcherpb.Event_ClientProgress:
			clientProgress.LoadPb(e.ClientProgress)
		default:
			return fmt.Errorf("unsupported batch fetcher event type: %T", e)
		}
		return nil
	})

	// The DeliverCert handler requests the transactions referenced by the received availability certificate
	// from the availability layer.
	// TODO: Make sure the certificate has been verified by the producer of this event.
	dsl.UponEvent[*eventpb.Event_DeliverCert](m, func(cert *eventpb.DeliverCert) error {
		// Create an empty output item and enqueue it immediately.
		// Actual output will be delayed until the transactions have been received.
		// This is necessary to preserve the order of incoming and outgoing events.
		item := outputItem{event: nil}
		output.Enqueue(&item)

		if cert.Cert.Type == nil {
			// Skip fetching transactions for padding certificates.
			// Directly deliver an empty batch instead.
			item.event = bfevents.NewOrderedBatch(mc.Destination, []*requestpb.Request{})
			output.Flush(m)
		} else {
			// If this is a proper certificate, request transactions from the availability layer.
			availabilitydsl.RequestTransactions(
				m,
				mc.Availability.Then(t.ModuleID(fmt.Sprintf("%v", epochNr))),
				cert.Cert,
				&txRequestContext{queueItem: &item},
			)
		}

		return nil
	})

	// The AppSnapshotRequest handler triggers a ClientProgress event (for the checkpointing protocol)
	// and forwards the original snapshot request event to the output.
	dsl.UponEvent[*eventpb.Event_AppSnapshotRequest](m, func(snapshotRequest *eventpb.AppSnapshotRequest) error {

		// Save the number of the epoch when the AppSnapshotRequest has been received.
		// This is necessary in case the epoch number changes
		// by the time the AppSnapshotRequest event is output and the hook function (added below) executed.
		epoch := epochNr

		// Forward the original event to the output.
		output.Enqueue(&outputItem{
			event: events.AppSnapshotRequest(mc.Destination, t.ModuleID(snapshotRequest.ReplyTo)),

			// At the time of forwarding, submit the client progress to the checkpointing protocol.
			f: func(_ *eventpb.Event) {
				clientProgress.GarbageCollect()
				dsl.EmitEvent(m, bfevents.ClientProgress(
					mc.Checkpoint.Then(t.ModuleID(fmt.Sprintf("%v", epoch))),
					clientProgress.Pb(),
				))
			},
		})
		output.Flush(m)

		return nil
	})

	// The ProvideTransactions handler filters the received transaction batch,
	// removing all transactions that have been previously delivered,
	// assigns the remaining transactions to the corresponding output item
	// (the one created on reception of the corresponding availability certificate in DeliverCert)
	// and flushes the output stream.
	availabilitydsl.UponProvideTransactions(m, func(txs []*requestpb.Request, context *txRequestContext) error {

		// Filter out transactions that already have been delivered
		newTxs := make([]*requestpb.Request, 0, len(txs))
		for _, req := range txs {
			// Runs for each received transaction.

			// Convenience variables
			clID := t.ClientID(req.ClientId)
			reqNo := t.ReqNo(req.ReqNo)

			// Only keep request if it has not yet been delivered.
			if clientProgress.Add(clID, reqNo) {
				newTxs = append(newTxs, req)
			}
		}

		context.queueItem.event = bfevents.NewOrderedBatch(mc.Destination, newTxs)
		output.Flush(m)
		return nil
	})

	// Explicitly ignore Init event. This prevents forwarding it to the destination module.
	dsl.UponEvent[*eventpb.Event_Init](m, func(_ *eventpb.Init) error {
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
