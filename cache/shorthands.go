package cache

import (
	"time"
)

// defaultObjectStorage is the object storage used by the package-level
// functions.
var (
	defaultObjectStorage *ObjectStorage
)

// DefaultObjectStorage returns the object storage on top of the default
// storage.
func DefaultObjectStorage() *ObjectStorage {
	if defaultObjectStorage == nil {
		defaultObjectStorage = ObjectStorageFor(DefaultStorage)
	}
	return defaultObjectStorage
}

// Get decodes a cached value from the default storage into obj.
// See ObjectStorage.Get.
func Get(key string, obj interface{}) error {
	return DefaultObjectStorage().Get(key, obj)
}

// Set caches a value in the default storage. See ObjectStorage.Set.
func Set(key string, obj interface{}, ttl time.Duration) error {
	return DefaultObjectStorage().Set(key, obj, ttl)
}

// Delete removes a cached value from the default storage.
func Delete(key string) error {
	return DefaultObjectStorage().Delete(key)
}
