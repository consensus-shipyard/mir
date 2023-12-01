// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package stats

import (
	"github.com/filecoin-project/mir/pkg/events"
	bfpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type StatInterceptor struct {
	*LiveStats

	// ID of the module that is consuming the transactions.
	// Statistics will only be performed on transactions destined to this module
	// and the rest of the events will be ignored by the StatInterceptor.
	txConsumerModule t.ModuleID
}

func NewStatInterceptor(s *LiveStats, txConsumer t.ModuleID) *StatInterceptor {
	return &StatInterceptor{s, txConsumer}
}

func (i *StatInterceptor) Intercept(events *events.EventList) (*events.EventList, error) {

	// Avoid nil dereference if Intercept is called on a nil *Recorder and simply do nothing.
	// This can happen if a pointer type to *Recorder is assigned to a variable with the interface type Interceptor.
	// Mir would treat that variable as non-nil, thinking there is an interceptor, and call Intercept() on it.
	// For more explanation, see https://mangatmodi.medium.com/go-check-nil-interface-the-right-way-d142776edef1
	if i == nil {
		return events, nil
	}

	it := events.Iterator()
	for evt := it.Next(); evt != nil; evt = it.Next() {

		switch e := evt.Type.(type) {
		case *eventpb.Event_Mempool:
			switch e := e.Mempool.Type.(type) {
			case *mempoolpb.Event_NewTransactions:
				for _, tx := range e.NewTransactions.Transactions {
					i.LiveStats.Submit(trantorpbtypes.TransactionFromPb(tx))
				}
			}
		case *eventpb.Event_BatchFetcher:

			// Skip events destined to other modules than the one consuming the transactions.
			if t.ModuleID(evt.DestModule) != i.txConsumerModule {
				continue
			}

			switch e := e.BatchFetcher.Type.(type) {
			case *bfpb.Event_NewOrderedBatch:
				for _, tx := range e.NewOrderedBatch.Txs {
					i.LiveStats.Deliver(trantorpbtypes.TransactionFromPb(tx))
				}
			}
		}
	}
	return events, nil
}
