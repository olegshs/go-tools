// Package memcached implements a cache driver for Memcached, using
// github.com/bradfitz/gomemcache. Keys are spread over the configured
// servers.
package memcached

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/olegshs/go-tools/cache/storage"
)

// Storage is the Memcached cache driver.
type Storage struct {
	client     *memcache.Client
	serverList *memcache.ServerList

	prefix string

	ttlDefault time.Duration
	ttlMax     time.Duration

	hits   int64
	misses int64
}

// NewStorage creates a Memcached storage. It panics if a server address is
// invalid. Connections are opened on demand.
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

// Get returns the entry for key, or storage.ErrNotFound if there is none.
// Memcached removes expired entries itself, so ErrExpired is not returned.
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

// Set stores the entry for key.
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

// Delete removes the entry for key. It returns memcache.ErrCacheMiss if
// there is none.
func (stor *Storage) Delete(key string) error {
	err := stor.client.Delete(stor.prefix + key)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAll flushes all servers, removing the entries of every storage and
// application that uses them.
func (stor *Storage) DeleteAll() error {
	err := stor.client.DeleteAll()
	if err != nil {
		return err
	}

	return nil
}

// Hits returns the number of successful Get calls.
func (stor *Storage) Hits() int64 {
	return atomic.LoadInt64(&stor.hits)
}

// Misses returns the number of Get calls that found no entry.
func (stor *Storage) Misses() int64 {
	return atomic.LoadInt64(&stor.misses)
}

// expire converts ttl to a Memcached expiration value: seconds for TTLs
// under 30 days, and a Unix timestamp for longer ones. Zero means never.
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
