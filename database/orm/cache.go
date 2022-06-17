package orm

import (
	"github.com/olegshs/go-tools/cache"
	"github.com/olegshs/go-tools/cache/drivers/blackhole"
)

var (
	cacheStorage = &cache.ObjectStorage{
		Storage: blackhole.NewStorage(),
	}
)

func SetCacheStorage(storage cache.StorageInterface) {
	cacheStorage = &cache.ObjectStorage{Storage: storage}
}

func SetCacheStorageByName(storageName string) {
	cacheStorage = cache.ObjectStorageFor(storageName)
}
