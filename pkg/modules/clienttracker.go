/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package modules

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
)

// TODO: Document this.

type ClientTracker interface {
	ApplyEvent(event *eventpb.Event) *events.EventList
	Status() (s *statuspb.ClientTrackerStatus, err error)
}
