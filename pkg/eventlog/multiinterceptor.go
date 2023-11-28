package eventlog

import (
	"github.com/filecoin-project/mir/pkg/events"
)

type repeater struct {
	interceptors []Interceptor
}

func (r *repeater) Intercept(events *events.EventList) (*events.EventList, error) {

	// Avoid nil dereference if Intercept is called on a nil *Recorder and simply do nothing.
	// This can happen if a pointer type to *Recorder is assigned to a variable with the interface type Interceptor.
	// Mir would treat that variable as non-nil, thinking there is an interceptor, and call Intercept() on it.
	// For more explanation, see https://mangatmodi.medium.com/go-check-nil-interface-the-right-way-d142776edef1
	if r == nil {
		return events, nil
	}

	for _, i := range r.interceptors {
		_, err := i.Intercept(events)
		if err != nil {
			return events, err
		}
	}
	return events, nil
}

func MultiInterceptor(interceptors ...Interceptor) Interceptor {
	return &repeater{interceptors: interceptors}
}
