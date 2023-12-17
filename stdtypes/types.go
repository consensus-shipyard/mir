package stdtypes

import (
	"github.com/multiformats/go-multiaddr"
)

// RetentionIndex represents an abstract notion of system progress used in garbage collection.
// The basic idea is to associate various parts of the system (parts of the state, even whole modules)
// that are subject to eventual garbage collection with a retention index.
// As the system progresses, the retention index monotonically increases
// and parts of the system associated with a lower retention index can be garbage-collected.
type RetentionIndex uint64

// Message represents a message to be sent over the network.
// It is the data type of the stdevents.SendMessage.Payload and stdevents.MessageReceived.Payload.
// The only requirement of a Message is that it must be serializable.
// Note that while a stdevents.SendMessage event as well as a stdevents.MessageReceived event can contain
// an arbitrary implementation of this interface in their Payload fields,
// a SendMessage or a MessageReceived event that underwent serialization and deserialization
// will always contain the RawMessage implementation of this interface,
// only holding the serialized representation of the original message.
// This is because Mir does not by itself know how to deserialize arbitrary (application-specific) messages.
type Message interface {
	ToBytes() ([]byte, error)
}

// NodeAddress represents the address of a node.
type NodeAddress multiaddr.Multiaddr
