// Package blackhole implements a cache driver that stores nothing. It is
// useful for disabling caching without changing code.
package blackhole

import (
	"sync/atomic"
	"time"

	"github.com/olegshs/go-tools/cache/storage"
)

// Storage is the blackhole driver. Every Get is a miss.
type Storage struct {
	misses int64
}

// NewStorage creates a blackhole storage.
func NewStorage() *Storage {
	stor := new(Storage)
	return stor
}

// Get always returns storage.ErrNotFound.
func (stor *Storage) Get(key string) ([]byte, error) {
	atomic.AddInt64(&stor.misses, 1)
	return nil, storage.ErrNotFound
}

// Set discards the data.
func (stor *Storage) Set(key string, data []byte, ttl time.Duration) error {
	return nil
}

// Delete always returns storage.ErrNotFound.
func (stor *Storage) Delete(key string) error {
	return storage.ErrNotFound
}

// DeleteAll does nothing.
func (stor *Storage) DeleteAll() error {
	return nil
}

// Hits always returns 0.
func (stor *Storage) Hits() int64 {
	return 0
}

// Misses returns the number of Get calls.
func (stor *Storage) Misses() int64 {
	return atomic.LoadInt64(&stor.misses)
}
