package storage

import (
	"time"

	"github.com/olegshs/go-tools/database"
)

// ConfigFile holds the settings of FileStorage. Dir is relative to
// config.BaseDir.
type ConfigFile struct {
	Dir string   `json:"dir"`
	GC  ConfigGC `json:"gc"`
}

// ConfigSql holds the settings of SqlStorage. Database is the name of a
// connection from the "database" config section. Encoding is "base64" or
// "none" (see package encoder); use "none" only with a binary column type.
type ConfigSql struct {
	Database string   `json:"database"`
	Table    string   `json:"table"`
	Encoding string   `json:"encoding"`
	GC       ConfigGC `json:"gc"`
}

// ConfigGC sets how often expired sessions are removed. Zero disables
// garbage collection.
type ConfigGC struct {
	Interval time.Duration `json:"interval"`
}

// DefaultConfigFile returns the default file storage settings.
func DefaultConfigFile() ConfigFile {
	return ConfigFile{
		Dir: "tmp/session",
		GC: ConfigGC{
			Interval: 0,
		},
	}
}

// DefaultConfigSql returns the default SQL storage settings.
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
