package cache

// Serializer is implemented by values that encode themselves for
// ObjectStorage instead of using gob.
type Serializer interface {
	Serialize() ([]byte, error)
}

// Deserializer is implemented by values that decode themselves from data
// written by their Serializer.
type Deserializer interface {
	Deserialize([]byte) error
}
