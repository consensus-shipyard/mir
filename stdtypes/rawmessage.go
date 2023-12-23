package stdtypes

// RawMessage is technically just raw data (RawData).
// We use a separate type to distinguish data that is intended to be sent/received over the network
// from data that is generated and used locally.
// The difference is important for making assumptions on the consistency / correctness of the data.
type RawMessage RawData

// ToBytes just delegates the call to the underlying type.
func (rm RawMessage) ToBytes() ([]byte, error) {
	return (RawData)(rm).ToBytes()
}
