package eventlog

import "github.com/filecoin-project/mir/pkg/events"

// Interceptor provides a way to gain insight into the internal operation of the node.
// Before being passed to the respective target modules, Events can be intercepted and logged
// for later analysis or replaying.
// An interceptor can be used for other purposes to, e.g., to gather statistical information or provide live monitoring.
type Interceptor interface {

	// Intercept is called by the node each Time Events are passed to a module.
	// ATTENTION: Since this is an interface type, it can happen that a nil value of a concrete type is used in the node
	// and, consequently, Intercept(events) will be called on nil.
	// The implementation of the concrete type must make sure that calling Intercept even on the nil value
	// does not cause any problems.
	// For more explanation, see https://mangatmodi.medium.com/go-check-nil-interface-the-right-way-d142776edef1
	Intercept(events *events.EventList) error
}
