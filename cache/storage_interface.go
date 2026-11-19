package cache

import (
	"time"
)

// StorageInterface is implemented by the cache drivers.
//
// Get returns storage.ErrNotFound or storage.ErrExpired for missing
// entries. Set stores data for ttl, which may be DefaultTTL or MaximumTTL.
// DeleteAll removes every entry; for Memcached and Redis this clears the
// whole server or database, not only this storage's keys. Hits and Misses
// count Get results since the storage was created.
type StorageInterface interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte, ttl time.Duration) error
	Delete(key string) error
	DeleteAll() error
	Hits() int64
	Misses() int64
}
