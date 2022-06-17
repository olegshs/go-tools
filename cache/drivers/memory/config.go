package memory

import (
	"time"

	"github.com/olegshs/go-tools/cache/storage"
)

type Config struct {
	storage.Config
	Size int64    `json:"size"`
	GC   ConfigGC `json:"gc"`
}

type ConfigGC struct {
	Interval time.Duration `json:"interval"`
}

func DefaultConfig() Config {
	return Config{
		Config: storage.DefaultConfig(),
		Size:   32,
		GC: ConfigGC{
			Interval: 60 * time.Second,
		},
	}
}
