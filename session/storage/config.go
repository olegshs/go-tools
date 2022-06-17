package storage

import (
	"time"

	"github.com/olegshs/go-tools/database"
)

type ConfigFile struct {
	Dir string   `json:"dir"`
	GC  ConfigGC `json:"gc"`
}

type ConfigSql struct {
	Database string   `json:"database"`
	Table    string   `json:"table"`
	Encoding string   `json:"encoding"`
	GC       ConfigGC `json:"gc"`
}

type ConfigGC struct {
	Interval time.Duration `json:"interval"`
}

func DefaultConfigFile() ConfigFile {
	return ConfigFile{
		Dir: "tmp/session",
		GC: ConfigGC{
			Interval: 0,
		},
	}
}

func DefaultConfigSql() ConfigSql {
	return ConfigSql{
		Database: database.DefaultDB,
		Table:    "sessions",
		Encoding: "base64",
		GC: ConfigGC{
			Interval: 0,
		},
	}
}
