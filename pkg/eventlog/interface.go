/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package eventlog

import "github.com/filecoin-project/mir/pkg/events"

// Interceptor provides a way to gain insight into the internal operation of the node.
// Before being passed to the respective target modules, events can be intercepted and logged
// for later analysis or replaying.
type Interceptor interface {

	// Intercept is called each time events are passed to a module, if an Interceptor is present in the node.
	// The expected behavior of Intercept is to add the intercepted events to a log for later analysis.
	// TODO: In the comment, also refer to the way events can be analyzed or replayed.
	Intercept(events *events.EventList) error
}
