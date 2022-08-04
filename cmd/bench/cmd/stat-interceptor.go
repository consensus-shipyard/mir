// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

type StatInterceptor struct {
	*Stats
}

func NewStatInterceptor(s *Stats) *StatInterceptor {
	return &StatInterceptor{s}
}

func (i *StatInterceptor) Intercept(events *events.EventList) error {
	it := events.Iterator()
	for e := it.Next(); e != nil; e = it.Next() {
		switch e := e.Type.(type) {
		case *eventpb.Event_NewRequests:
			for _, req := range e.NewRequests.Requests {
				i.Stats.NewRequest(req)
			}
		case *eventpb.Event_Deliver:
			for _, req := range e.Deliver.Batch.Requests {
				i.Stats.Delivered(req.Req)
			}
		}
	}
	return nil
}
