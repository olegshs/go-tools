// Package redis implements a cache driver for Redis, using
// github.com/gomodule/redigo.
package redis

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/olegshs/go-tools/cache/storage"
)

// Storage is the Redis cache driver.
type Storage struct {
	pool *redis.Pool

	prefix string

	ttlDefault time.Duration
	ttlMax     time.Duration

	hits   int64
	misses int64
}

// NewStorage creates a Redis storage with a connection pool. Connections
// are opened on demand.
func NewStorage(conf Config) *Storage {
	stor := new(Storage)
	stor.pool = &redis.Pool{
		Dial: func() (conn redis.Conn, err error) {
			address := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
			options := []redis.DialOption{
				redis.DialDatabase(conf.DB),
				redis.DialPassword(conf.Password),
			}
			return redis.Dial("tcp", address, options...)
		},
		MaxIdle:         conf.Pool.MaxIdle,
		MaxActive:       conf.Pool.MaxActive,
		IdleTimeout:     conf.Pool.IdleTimeout,
		Wait:            conf.Pool.Wait,
		MaxConnLifetime: conf.Pool.MaxConnLifetime,
	}

	stor.prefix = conf.Prefix

	stor.ttlDefault = conf.TTL.Default
	stor.ttlMax = conf.TTL.Maximum

	return stor
}

// Get returns the entry for key, or storage.ErrNotFound if there is none.
// Redis removes expired entries itself, so ErrExpired is not returned.
func (stor *Storage) Get(key string) ([]byte, error) {
	conn := stor.pool.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", stor.prefix+key))
	if err == redis.ErrNil {
		atomic.AddInt64(&stor.misses, 1)
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	atomic.AddInt64(&stor.hits, 1)

	return b, nil
}

// Set stores the entry for key, with a PX expiry if ttl is positive.
func (stor *Storage) Set(key string, data []byte, ttl time.Duration) error {
	conn := stor.pool.Get()
	defer conn.Close()

	args := []interface{}{
		stor.prefix + key,
		data,
	}

	expire := stor.expire(ttl)
	if expire > 0 {
		args = append(args, "PX", expire)
	}

	_, err := conn.Do("SET", args...)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes the entry for key. Missing keys are not an error.
func (stor *Storage) Delete(key string) error {
	conn := stor.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", stor.prefix+key)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAll runs FLUSHDB, removing every key in the Redis database, not
// only this storage's keys.
func (stor *Storage) DeleteAll() error {
	conn := stor.pool.Get()
	defer conn.Close()

	_, err := conn.Do("FLUSHDB")
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

// expire returns ttl in milliseconds, after applying the default and
// maximum TTL. Zero or less means no expiry.
func (stor *Storage) expire(ttl time.Duration) int64 {
	if ttl == storage.DefaultTTL {
		ttl = stor.ttlDefault
	}

	if stor.ttlMax > 0 && (ttl > stor.ttlMax || ttl == storage.MaximumTTL) {
		ttl = stor.ttlMax
	}

	return int64(ttl / time.Millisecond)
}
