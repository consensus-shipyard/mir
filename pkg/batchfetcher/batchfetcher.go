package batchfetcher

import (
	"fmt"

	availabilitydsl "github.com/filecoin-project/mir/pkg/availability/dsl"
	availabilityevents "github.com/filecoin-project/mir/pkg/availability/events"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// NewModule returns a new batch fetcher module.
// The batch receives events output by the ordering protocol (e.g. ISS)
// and relays them to the application in the same order.
// It replaces the DeliverCert events from the input stream by the corresponding ProvideTransactions
// that it obtains from the availability layer.
// It keeps track of the current epoch (by observing the relayed NewEpoch events)
// and automatically requests the transactions from the correct instance of the availability module.
func NewModule(mc *ModuleConfig) modules.Module {
	m := dsl.NewModule(mc.Self)

	// The current epoch number, as announced by the ordering protocol.
	epochNr := t.EpochNr(0)

	// Map of delivered requests that is used to filter duplicates.
	// TODO: Implement compaction (client watermarks) so that this map does not grow indefinitely.
	delivered := make(map[t.ClientID]map[t.ReqNo]struct{})

	// Queue of output events. It is required for buffering events being relayed
	// in case a DeliverCert event that has arrived before has not yet been transformed to a ProvideTransactions event.
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

	// The DeliverCert handler requests the transactions referenced by the received availability certificate
	// from the availability layer.
	// TODO: Verify Certificate? Here or elsewhere?
	dsl.UponEvent[*eventpb.Event_DeliverCert](m, func(cert *eventpb.DeliverCert) error {
		// Skip padding certificates. DeliverCert events with nil certificates are considered noops.
		if cert.Cert.Type == nil {
			return nil
		}

		switch c := cert.Cert.Type.(type) {
		case *availabilitypb.Cert_Msc:

			// TODO: Check whether this makes any sense.
			if len(c.Msc.BatchId) == 0 {
				fmt.Println("Received empty batch availability certificate.")
				return nil
			}

			// Create an empty output item and enqueue it immediately.
			// Actual output will be delayed until the transactions have been received.
			// This is necessary to preserve the order of incoming and outgoing events.
			item := outputItem{event: nil}
			output.Enqueue(&item)

			// Request transactions from the availability layer.
			availabilitydsl.RequestTransactions(
				m,
				mc.Availability.Then(t.ModuleID(fmt.Sprintf("%v", epochNr))),
				cert.Cert,
				&item,
			)

		default:
			return fmt.Errorf("unknown availability certificate type: %T", cert.Cert.Type)
		}
		return nil
	})

	// The ProvideTransactions handler filters the received transaction batch,
	// removing all transactions that have been previously delivered,
	// assigns the remaining transactions to the corresponding output item
	// (the one created on reception of the corresponding availability certificate in DeliverCert)
	// and flushes the output stream.
	availabilitydsl.UponProvideTransactions(m, func(txs []*requestpb.Request, item *outputItem) error {

		// Filter out transactions that already have been delivered
		newTxs := make([]*requestpb.Request, 0, len(txs))
		for _, req := range txs {
			// Runs for each received transaction.

			// Convenience variables
			clID := t.ClientID(req.ClientId)
			reqNo := t.ReqNo(req.ReqNo)

			// Only keep request if it has not yet been delivered.
			// TODO: Make this more efficient by compacting the delivered set.
			_, ok := delivered[clID]
			if !ok {
				// If this is the first transaction from this client, create a new entry for the ClientID.
				delivered[clID] = make(map[t.ReqNo]struct{})
			}
			if _, ok := delivered[clID][reqNo]; !ok {
				// If the transaction has not yet been delivered, record its delivery and append it to the output.
				delivered[clID][reqNo] = struct{}{}
				newTxs = append(newTxs, req)
			}
		}

		item.event = availabilityevents.ProvideTransactions(mc.Destination, newTxs, nil)
		output.Flush(m)
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
