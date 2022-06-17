// Пакет redis реализует драйвер для работы с Redis.
package redis

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/olegshs/go-tools/cache/storage"
)

type Storage struct {
	pool *redis.Pool

	prefix string

	ttlDefault time.Duration
	ttlMax     time.Duration

	hits   int64
	misses int64
}

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

func (stor Storage) Get(key string) ([]byte, error) {
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

func (stor Storage) Set(key string, data []byte, ttl time.Duration) error {
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

func (stor Storage) Delete(key string) error {
	conn := stor.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", stor.prefix+key)
	if err != nil {
		return err
	}

	return nil
}

func (stor Storage) DeleteAll() error {
	conn := stor.pool.Get()
	defer conn.Close()

	_, err := conn.Do("FLUSHDB")
	if err != nil {
		return err
	}

	return nil
}

func (stor Storage) Hits() int64 {
	return stor.hits
}

func (stor Storage) Misses() int64 {
	return stor.misses
}

func (stor Storage) expire(ttl time.Duration) int64 {
	if ttl == storage.DefaultTTL {
		ttl = stor.ttlDefault
	}

	if stor.ttlMax > 0 && (ttl > stor.ttlMax || ttl == storage.MaximumTTL) {
		ttl = stor.ttlMax
	}

	return int64(ttl / time.Millisecond)
}
