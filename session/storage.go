package session

import (
	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/session/storage"
)

// newStorage creates the storage named by conf.Storage.Driver from the
// "session.storage" config section. It returns nil for unknown drivers.
func newStorage(conf Config) StorageInterface {
	var stor StorageInterface
	confKey := "session.storage"

	switch conf.Storage.Driver {
	case storage.StorageFile:
		conf := storage.DefaultConfigFile()
		config.GetStruct(confKey, &conf)
		stor = storage.NewFileStorage(conf)

	case storage.StorageSql:
		conf := storage.DefaultConfigSql()
		config.GetStruct(confKey, &conf)
		stor = storage.NewSqlStorage(conf)

	default:
		return nil
	}

	stor.SetTTL(conf.TTL)

	return stor
}
