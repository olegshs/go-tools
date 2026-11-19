// Package cache provides named cache storages with interchangeable drivers:
// in-process memory, files, Memcached, Redis, and a no-op "blackhole".
//
// A storage is configured under "cache.<name>". Storages without their own
// section use "cache.default". The driver-specific settings sit in the same
// section as the driver name:
//
//	cache:
//	  default:
//	    driver: redis
//	    host: localhost
//	    ttl:
//	      default: 10m
//
// StorageInterface stores raw bytes. ObjectStorage and the package-level
// Get, Set and Delete functions store Go values encoded with gob, or with
// the value's own Serializer and Deserializer methods.
package cache

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/olegshs/go-tools/cache/drivers/blackhole"
	"github.com/olegshs/go-tools/cache/drivers/file"
	"github.com/olegshs/go-tools/cache/drivers/memcached"
	"github.com/olegshs/go-tools/cache/drivers/memory"
	"github.com/olegshs/go-tools/cache/drivers/redis"
	"github.com/olegshs/go-tools/cache/storage"
	"github.com/olegshs/go-tools/config"
)

const (
	// DefaultStorage is the name of the storage used by the package-level
	// functions and the fallback config section for unconfigured storages.
	DefaultStorage = "default"

	// Driver names. An empty driver means DriverMemory.
	DriverBlackhole = "blackhole"
	DriverFile      = "file"
	DriverMemory    = "memory"
	DriverMemcached = "memcached"
	DriverRedis     = "redis"

	// Special TTL values for Set. DefaultTTL uses the storage's configured
	// default TTL. MaximumTTL uses the configured maximum TTL, or no expiry if
	// there is no maximum.
	DefaultTTL = storage.DefaultTTL
	MaximumTTL = storage.MaximumTTL
)

var (
	// storageInstances caches storages by name.
	storageInstances      = map[string]StorageInterface{}
	storageInstancesMutex sync.Mutex
)

// Storage returns the storage with the given name, creating it from config
// on first use. It panics if the driver is unknown or cannot be set up.
func Storage(name string) StorageInterface {
	storageInstancesMutex.Lock()
	defer storageInstancesMutex.Unlock()

	stor, ok := storageInstances[name]
	if ok {
		return stor
	}

	stor = newStorage(name)
	storageInstances[name] = stor

	return stor
}

// newStorage creates a storage from config and starts its statistics log if
// it is enabled.
func newStorage(name string) StorageInterface {
	conf := storage.DefaultConfig()
	getStorageConfig(name, &conf)

	stor := newStorageByDriver(conf.Driver, name)

	if conf.Log.Enabled {
		log := NewLog(stor, conf.Log)
		log.Start()
	}

	return stor
}

// newStorageByDriver creates a storage with the given driver. Storages of
// the same driver are kept apart: the file driver uses a subdirectory named
// after the storage, and Memcached and Redis prefix keys with "(name)".
func newStorageByDriver(driver string, name string) StorageInterface {
	switch driver {
	case DriverBlackhole:
		return blackhole.NewStorage()

	case DriverFile:
		conf := file.DefaultConfig()
		getStorageConfig(name, &conf)
		conf.Dir = filepath.Join(conf.Dir, name)
		return file.NewStorage(conf)

	case DriverMemory, "":
		conf := memory.DefaultConfig()
		getStorageConfig(name, &conf)
		return memory.NewStorage(conf)

	case DriverMemcached:
		conf := memcached.DefaultConfig()
		getStorageConfig(name, &conf)
		conf.Prefix = addPrefix(conf.Prefix, name)
		return memcached.NewStorage(conf)

	case DriverRedis:
		conf := redis.DefaultConfig()
		getStorageConfig(name, &conf)
		conf.Prefix = addPrefix(conf.Prefix, name)
		return redis.NewStorage(conf)

	default:
		panic("invalid cache driver: " + driver)
	}
}

// getStorageConfig fills conf from "cache.<name>", or from "cache.default"
// if that section does not exist, and reports whether the named section
// exists.
func getStorageConfig(name string, conf interface{}) (exists bool) {
	const prefix = "cache."
	key := prefix + name

	exists = config.Exists(key)
	if !exists {
		key = prefix + DefaultStorage
	}

	config.GetStruct(key, conf)
	return
}

// addPrefix returns s with "(prefix)" in front of it.
func addPrefix(s, prefix string) string {
	return fmt.Sprintf("(%s)%s", prefix, s)
}
