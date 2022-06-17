// Пакет blackhole реализует фиктивный драйвер, данные не сохраняются.
package blackhole

import (
	"sync/atomic"
	"time"

	"github.com/olegshs/go-tools/cache/storage"
)

type Storage struct {
	misses int64
}

func NewStorage() *Storage {
	stor := new(Storage)
	return stor
}

func (stor *Storage) Get(key string) ([]byte, error) {
	atomic.AddInt64(&stor.misses, 1)
	return nil, storage.ErrNotFound
}

func (stor *Storage) Set(key string, data []byte, ttl time.Duration) error {
	return nil
}

func (stor *Storage) Delete(key string) error {
	return storage.ErrNotFound
}

func (stor *Storage) DeleteAll() error {
	return nil
}

func (stor *Storage) Hits() int64 {
	return 0
}

func (stor *Storage) Misses() int64 {
	return stor.misses
}
