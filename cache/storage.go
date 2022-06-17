// Пакет cache предоставляет средства для кэширования данных.
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
	DefaultStorage = "default"

	DriverBlackhole = "blackhole"
	DriverFile      = "file"
	DriverMemory    = "memory"
	DriverMemcached = "memcached"
	DriverRedis     = "redis"

	DefaultTTL = storage.DefaultTTL
	MaximumTTL = storage.MaximumTTL
)

var (
	storageInstances      = map[string]StorageInterface{}
	storageInstancesMutex sync.Mutex
)

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

func addPrefix(s, prefix string) string {
	return fmt.Sprintf("(%s)%s", prefix, s)
}
