// Пакет memcached реализует драйвер для работы с Memcached.
package memcached

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/olegshs/go-tools/cache/storage"
)

type Storage struct {
	client     *memcache.Client
	serverList *memcache.ServerList

	prefix string

	ttlDefault time.Duration
	ttlMax     time.Duration

	hits   int64
	misses int64
}

func NewStorage(conf Config) *Storage {
	stor := new(Storage)

	stor.serverList = new(memcache.ServerList)
	for _, server := range conf.Servers {
		s := fmt.Sprintf("%s:%d", server.Host, server.Port)

		err := stor.serverList.SetServers(s)
		if err != nil {
			panic(err)
		}
	}

	stor.client = memcache.NewFromSelector(stor.serverList)

	stor.prefix = conf.Prefix

	stor.ttlDefault = conf.TTL.Default
	stor.ttlMax = conf.TTL.Maximum

	return stor
}

func (stor *Storage) Get(key string) ([]byte, error) {
	item, err := stor.client.Get(stor.prefix + key)
	if err == memcache.ErrCacheMiss {
		atomic.AddInt64(&stor.misses, 1)
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	atomic.AddInt64(&stor.hits, 1)

	return item.Value, nil
}

func (stor *Storage) Set(key string, data []byte, ttl time.Duration) error {
	item := &memcache.Item{
		Key:        stor.prefix + key,
		Value:      data,
		Expiration: stor.expire(ttl),
	}

	err := stor.client.Set(item)
	if err != nil {
		return err
	}

	return nil
}

func (stor *Storage) Delete(key string) error {
	err := stor.client.Delete(stor.prefix + key)
	if err != nil {
		return err
	}

	return nil
}

func (stor *Storage) DeleteAll() error {
	err := stor.client.DeleteAll()
	if err != nil {
		return err
	}

	return nil
}

func (stor *Storage) Hits() int64 {
	return stor.hits
}

func (stor *Storage) Misses() int64 {
	return stor.misses
}

func (stor *Storage) expire(ttl time.Duration) int32 {
	if ttl == storage.DefaultTTL {
		ttl = stor.ttlDefault
	}

	if stor.ttlMax > 0 && (ttl > stor.ttlMax || ttl == storage.MaximumTTL) {
		ttl = stor.ttlMax
	}

	// https://github.com/memcached/memcached/wiki/Programming#expiration
	if ttl < 60*60*24*30*time.Second {
		return int32(ttl / time.Second)
	}

	return int32(time.Now().Add(ttl).Unix())
}
