package events

import t "github.com/filecoin-project/mir/pkg/types"

type Event interface {

	// Src returns the module that emitted the event.
	// While this information is not always necessary for the system operation,
	// it is useful for analyzing event traces and debugging.
	Src() t.ModuleID

	// Dest returns the destination module of the event.
	Dest() t.ModuleID

	// Bytes returns a serialized representation of the event
	// as a slice of bytes from which the event can be reconstructed.
	// Note that Bytes does not necessarily guarantee the output to be deterministic.
	// Even multiple subsequent calls to Bytes on the same event object might return different byte slices.
	// If an error occurs during serialization, Bytes returns a nil byte slice and a non-nil error.
	Bytes() ([]byte, error)

	// String returns a human-readable representation of the event.
	// While not used by the runtime itself, it can be used by associated tools.
	// Conventionally, this representation is JSON.
	String() string
}
