package rendezvous

import "context"

// The Syncer serves the purpose of synchronizing multiple processes in time.
// Each process declares when it is ready to proceed and blocks. Only after all processes have declared readiness,
// all processes are released.
type Syncer interface {

	// Ready blocks until all other processes have called Ready() and then returns nil.
	// Ready returns an error when ctx is canceled or an error occurs.
	Ready(ctx context.Context) error
}
