package orm

import (
	"github.com/olegshs/go-tools/cache"
	"github.com/olegshs/go-tools/cache/drivers/blackhole"
)

var (
	// cacheStorage caches loaded models. The default blackhole storage
	// disables caching.
	cacheStorage = &cache.ObjectStorage{
		Storage: blackhole.NewStorage(),
	}
)

// SetCacheStorage enables model caching in storage. Models are cached
// with gob, so their field types must be gob-encodable. Changes made
// without the ORM are not seen until the entries expire, unless
// DeleteFromCache is called.
func SetCacheStorage(storage cache.StorageInterface) {
	cacheStorage = &cache.ObjectStorage{Storage: storage}
}

// SetCacheStorageByName enables model caching in the named cache storage.
// See SetCacheStorage.
func SetCacheStorageByName(storageName string) {
	cacheStorage = cache.ObjectStorageFor(storageName)
}
