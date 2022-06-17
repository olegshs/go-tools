// Пакет memory реализует драйвер для хранения данных в памяти приложения.
package memory

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olegshs/go-tools/cache/storage"
	"github.com/olegshs/go-tools/helpers"
)

type Storage struct {
	items      map[string]*storageItem
	itemsMutex sync.RWMutex

	ttlDefault time.Duration
	ttlMax     time.Duration

	gc *helpers.Interval

	size    int64
	sizeMax int64

	hits   int64
	misses int64
}

func NewStorage(conf Config) *Storage {
	stor := new(Storage)

	stor.items = map[string]*storageItem{}

	stor.ttlDefault = conf.TTL.Default
	stor.ttlMax = conf.TTL.Maximum

	if conf.GC.Interval > 0 {
		stor.gc = helpers.NewInterval(conf.GC.Interval, stor.gcRun)
		stor.gc.RandomDelay(conf.GC.Interval / 8)
		stor.gc.Start()
	}

	stor.sizeMax = conf.Size * 1024 * 1024

	return stor
}

func (stor *Storage) Get(key string) ([]byte, error) {
	stor.itemsMutex.RLock()
	defer stor.itemsMutex.RUnlock()

	item, ok := stor.items[key]
	if !ok {
		atomic.AddInt64(&stor.misses, 1)
		return nil, storage.ErrNotFound
	}

	if item.isExpired() {
		atomic.AddInt64(&stor.misses, 1)
		return nil, storage.ErrExpired
	}

	atomic.AddInt64(&stor.hits, 1)
	atomic.AddInt64(&item.hits, 1)

	return item.data, nil
}

func (stor *Storage) Set(key string, data []byte, ttl time.Duration) error {
	if stor.sizeMax > 0 && int64(len(data)) > stor.sizeMax {
		return storage.ErrInvalidData
	}

	stor.itemsMutex.Lock()
	defer stor.itemsMutex.Unlock()

	item, ok := stor.items[key]
	if ok {
		stor.size -= item.size()
	} else {
		item = new(storageItem)
		item.created = time.Now().Unix()
	}

	item.expire = storage.Expire(ttl, stor.ttlDefault, stor.ttlMax)
	item.data = data

	itemSize := item.size()

	if stor.sizeMax > 0 && stor.size+itemSize > stor.sizeMax {
		size1 := stor.sizeMax - itemSize
		size2 := stor.sizeMax / 2
		if size1 < size2 {
			stor.truncate(size1)
		} else {
			stor.truncate(size2)
		}
	}

	stor.items[key] = item
	stor.size += itemSize

	return nil
}

func (stor *Storage) Delete(key string) error {
	stor.itemsMutex.Lock()
	defer stor.itemsMutex.Unlock()

	item, ok := stor.items[key]
	if !ok {
		return storage.ErrNotFound
	}

	stor.size -= item.size()
	delete(stor.items, key)

	return nil
}

func (stor *Storage) DeleteAll() error {
	stor.itemsMutex.Lock()
	defer stor.itemsMutex.Unlock()

	stor.items = map[string]*storageItem{}
	stor.size = 0

	return nil
}

func (stor *Storage) Hits() int64 {
	return stor.hits
}

func (stor *Storage) Misses() int64 {
	return stor.misses
}

func (stor *Storage) gcRun() {
	stor.itemsMutex.Lock()
	defer stor.itemsMutex.Unlock()

	for key, item := range stor.items {
		if item.isExpired() {
			stor.size -= item.size()
			delete(stor.items, key)
		}
	}
}

func (stor *Storage) truncate(size int64) {
	if stor.size <= size {
		return
	}

	if size <= 0 {
		stor.size = 0
		stor.items = map[string]*storageItem{}
		return
	}

	index := newIndex(stor.items)
	sort.Sort(index)

	for _, indexItem := range index {
		item := stor.items[indexItem.key]

		stor.size -= item.size()
		delete(stor.items, indexItem.key)

		if stor.size <= size {
			return
		}
	}
}
