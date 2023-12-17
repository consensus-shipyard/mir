package stdtypes

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
