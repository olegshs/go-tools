package file

import (
	"time"

	"github.com/olegshs/go-tools/cache/storage"
)

// Config holds the file driver settings. Dir is relative to config.BaseDir;
// each storage uses a subdirectory named after it. Hash names the function
// used to turn keys into file names.
type Config struct {
	storage.Config
	Dir  string   `json:"dir"`
	Hash string   `json:"hash"`
	GC   ConfigGC `json:"gc"`
}

// ConfigGC sets how often expired files are removed. Zero disables garbage
// collection, and expired files are then only detected on read.
type ConfigGC struct {
	Interval time.Duration `json:"interval"`
}

// DefaultConfig returns the default file driver settings.
func DefaultConfig() Config {
	return Config{
		Config: storage.DefaultConfig(),
		Dir:    "tmp/cache",
		Hash:   "md5",
		GC: ConfigGC{
			Interval: 0,
		},
	}
}
