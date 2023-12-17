package stdtypes

import (
	"fmt"
	"strconv"
)

// RetentionIndex represents an abstract notion of system progress used in garbage collection.
// The basic idea is to associate various parts of the system (parts of the state, even whole modules)
// that are subject to eventual garbage collection with a retention index.
// As the system progresses, the retention index monotonically increases
// and parts of the system associated with a lower retention index can be garbage-collected.
type RetentionIndex uint64

// Message represents a message to be sent over the network.
// It is the data type of the SendMessage.Payload and MessageReceived.Payload.
// The only requirement of a Message is that it must be serializable.
// Note that while a SendMessage event as well as a MessageReceived event can contain
// an arbitrary implementation of this interface in their Payload fields,
// A SendMessage or a MessageReceived event that underwent serialization and deserialization
// will always contain the RawMessage implementation of this interface,
// only holding the serialized representation of the original message.
// This is because Mir does not by itself know how to deserialize arbitrary (application-specific) messages.
type Message interface {
	ToBytes() ([]byte, error)
}

type RawMessage []byte

func (m RawMessage) ToBytes() ([]byte, error) {
	return m, nil
}

func (m RawMessage) String() string {
	if isPrintableString(m) {
		return string(m)
	}
	return fmt.Sprintf("%v", []byte(m))
}

func isPrintableString(data []byte) bool {
	for _, r := range string(data) {
		if !strconv.IsPrint(r) {
			return false
		}
	}
	return true
	// TODO: Replace the above by the commented code below once the appropriate version of Go is supported.
	//   Currently we do not support such a high version of Go, because we are using an old version of libp2p
	//   that requires an old Go version (tested with go 1.19). However, updating libp2p to a higher version
	//   breaks the libp2p transport module. That needs to be fixed first.
	//return !bytes.ContainsFunc(data, func(r rune) bool {
	//	return !strconv.IsPrint(r)
	//})
}
