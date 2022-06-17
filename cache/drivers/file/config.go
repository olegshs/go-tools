package file

import (
	"time"

	"github.com/olegshs/go-tools/cache/storage"
)

type Config struct {
	storage.Config
	Dir  string   `json:"dir"`
	Hash string   `json:"hash"`
	GC   ConfigGC `json:"gc"`
}

type ConfigGC struct {
	Interval time.Duration `json:"interval"`
}

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
