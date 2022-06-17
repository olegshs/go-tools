package cache

import (
	"time"
)

var (
	defaultObjectStorage *ObjectStorage
)

func DefaultObjectStorage() *ObjectStorage {
	if defaultObjectStorage == nil {
		defaultObjectStorage = ObjectStorageFor(DefaultStorage)
	}
	return defaultObjectStorage
}

func Get(key string, obj interface{}) error {
	return DefaultObjectStorage().Get(key, obj)
}

func Set(key string, obj interface{}, ttl time.Duration) error {
	return DefaultObjectStorage().Set(key, obj, ttl)
}

func Delete(key string) error {
	return DefaultObjectStorage().Delete(key)
}
