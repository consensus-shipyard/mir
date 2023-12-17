// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package stats

import (
	bfpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/stdtypes"
)

type StatInterceptor struct {
	*LiveStats

	// ID of the module that is consuming the transactions.
	// Statistics will only be performed on transactions destined to this module
	// and the rest of the events will be ignored by the StatInterceptor.
	txConsumerModule stdtypes.ModuleID
}

func NewStatInterceptor(s *LiveStats, txConsumer stdtypes.ModuleID) *StatInterceptor {
	return &StatInterceptor{s, txConsumer}
}

func (i *StatInterceptor) Intercept(events *stdtypes.EventList) error {

	// Avoid nil dereference if Intercept is called on a nil *Recorder and simply do nothing.
	// This can happen if a pointer type to *Recorder is assigned to a variable with the interface type Interceptor.
	// Mir would treat that variable as non-nil, thinking there is an interceptor, and call Intercept() on it.
	// For more explanation, see https://mangatmodi.medium.com/go-check-nil-interface-the-right-way-d142776edef1
	if i == nil {
		return nil
	}

	it := events.Iterator()
	for evt := it.Next(); evt != nil; evt = it.Next() {

		// Skip events of unknown types.
		pbevt, ok := evt.(*eventpb.Event)
		if !ok {
			continue
		}

		switch e := pbevt.Type.(type) {
		case *eventpb.Event_Mempool:
			switch e := e.Mempool.Type.(type) {
			case *mempoolpb.Event_NewTransactions:
				for _, tx := range e.NewTransactions.Transactions {
					i.LiveStats.Submit(trantorpbtypes.TransactionFromPb(tx))
				}
			}
		case *eventpb.Event_BatchFetcher:

			// Skip events destined to other modules than the one consuming the transactions.
			if evt.Dest() != i.txConsumerModule {
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
	return nil
}
