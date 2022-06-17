package cache

type Serializer interface {
	Serialize() ([]byte, error)
}

type Deserializer interface {
	Deserialize([]byte) error
}
