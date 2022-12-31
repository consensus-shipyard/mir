package eventlog

import (
	"github.com/filecoin-project/mir/pkg/events"
)

type repeater struct {
	interceptors []Interceptor
}

func (r *repeater) Intercept(events *events.EventList) error {
	for _, i := range r.interceptors {
		if err := i.Intercept(events); err != nil {
			return err
		}
	}
	return nil
}

func MultiInterceptor(interceptors ...Interceptor) Interceptor {
	return &repeater{interceptors: interceptors}
}
