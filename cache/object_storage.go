package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"
)

var (
	// ErrEmptyObject is returned by ObjectStorage.Get when the entry was stored
	// with a nil value.
	ErrEmptyObject = errors.New("empty object")
)

// ObjectStorage stores Go values in a StorageInterface. Values are encoded
// with gob unless they implement Serializer and Deserializer. Keys get an
// ".(obj)" suffix, so they do not collide with raw entries in the same
// storage.
type ObjectStorage struct {
	Storage StorageInterface
}

// ObjectStorageFor returns an ObjectStorage on top of the named storage.
func ObjectStorageFor(name string) *ObjectStorage {
	return &ObjectStorage{Storage(name)}
}

// Get decodes the entry for key into obj, which must be a pointer or a
// Deserializer.
func (storage *ObjectStorage) Get(key string, obj interface{}) error {
	key = objectStorageKey(key)

	data, err := storage.Storage.Get(key)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return ErrEmptyObject
	}

	switch t := obj.(type) {
	case Deserializer:
		err := t.Deserialize(data)
		if err != nil {
			return err
		}
	default:
		dec := gob.NewDecoder(bytes.NewBuffer(data))
		err := dec.Decode(obj)
		if err != nil {
			return err
		}
	}

	return nil
}

// Set encodes obj and stores it under key for ttl. A nil obj stores an
// empty entry, which Get reports as ErrEmptyObject.
func (storage *ObjectStorage) Set(key string, obj interface{}, ttl time.Duration) error {
	buf := new(bytes.Buffer)

	switch t := obj.(type) {
	case Serializer:
		data, err := t.Serialize()
		if err != nil {
			return err
		}
		buf.Write(data)
	case nil:
		// empty
	default:
		enc := gob.NewEncoder(buf)
		err := enc.Encode(obj)
		if err != nil {
			return err
		}
	}

	key = objectStorageKey(key)

	err := storage.Storage.Set(key, buf.Bytes(), ttl)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes the entry for key.
func (storage *ObjectStorage) Delete(key string) error {
	key = objectStorageKey(key)

	err := storage.Storage.Delete(key)
	if err != nil {
		return err
	}

	return nil
}

func objectStorageKey(key string) string {
	return key + ".(obj)"
}
