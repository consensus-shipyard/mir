package net

// Stats represents a statistics tracker for networking.
// It tracks the amount of data that is sent, received, and dropped (i.e., not sent).
// The data can be categorized using string labels, and it is the implementation's choice how to interpret them.
// All method implementations must be thread-safe, as they can be called concurrently.
type Stats interface {
	Sent(nBytes int, label string)
	Received(nBytes int, label string)
	Dropped(nBytes int, label string)
}
