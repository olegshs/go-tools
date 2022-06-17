package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"
)

var (
	ErrEmptyObject = errors.New("empty object")
)

type ObjectStorage struct {
	Storage StorageInterface
}

func ObjectStorageFor(name string) *ObjectStorage {
	return &ObjectStorage{Storage(name)}
}

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
