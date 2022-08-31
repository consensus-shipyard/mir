// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/filecoin-project/mir/pkg/events"
	bfpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type StatInterceptor struct {
	*Stats

	// ID of the module that is consuming the transactions.
	// Statistics will only be performed on transactions destined to this module
	// and the rest of the events will be ignored by the StatInterceptor.
	txConsumerModule t.ModuleID
}

func NewStatInterceptor(s *Stats, txConsumer t.ModuleID) *StatInterceptor {
	return &StatInterceptor{s, txConsumer}
}

func (i *StatInterceptor) Intercept(events *events.EventList) error {
	it := events.Iterator()
	for e := it.Next(); e != nil; e = it.Next() {

		// Skip events destined to other modules than the one consuming the transactions.
		if t.ModuleID(e.DestModule) != i.txConsumerModule {
			continue
		}

		switch e := e.Type.(type) {
		case *eventpb.Event_NewRequests:
			for _, req := range e.NewRequests.Requests {
				i.Stats.NewRequest(req)
			}
		case *eventpb.Event_BatchFetcher:
			switch e := e.BatchFetcher.Type.(type) {
			case *bfpb.Event_NewOrderedBatch:
				for _, req := range e.NewOrderedBatch.Txs {
					i.Stats.Delivered(req)
				}
			}
		}
	}
	return nil
}
