package cache

import (
	"time"
)

type StorageInterface interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte, ttl time.Duration) error
	Delete(key string) error
	DeleteAll() error
	Hits() int64
	Misses() int64
}
