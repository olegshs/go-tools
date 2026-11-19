package memory

import (
	"time"

	"github.com/olegshs/go-tools/cache/storage"
)

// Config holds the memory driver settings. Size is the maximum total data
// size in MiB; zero means no limit.
type Config struct {
	storage.Config
	Size int64    `json:"size"`
	GC   ConfigGC `json:"gc"`
}

// ConfigGC sets how often expired entries are removed. Zero disables
// garbage collection.
type ConfigGC struct {
	Interval time.Duration `json:"interval"`
}

// DefaultConfig returns the default memory driver settings: a 32 MiB limit
// and garbage collection every minute.
func DefaultConfig() Config {
	return Config{
		Config: storage.DefaultConfig(),
		Size:   32,
		GC: ConfigGC{
			Interval: 60 * time.Second,
		},
	}
}
