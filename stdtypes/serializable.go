package stdtypes

// Serializable is the type used for certain application-specific objects included in some standard events (stdevents).
// For example, the parameters to a submodule are required to be serializable.
// When deserializing events from the stdevents package, such objects will deserialize to RawData.
type Serializable interface {
	ToBytes() ([]byte, error)
}
