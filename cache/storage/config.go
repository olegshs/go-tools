// Пакет storage содержит общие для различных драйверов функции и определения.
package storage

import (
	"time"

	"github.com/olegshs/go-tools/logs"
)

type Config struct {
	Driver string    `json:"driver"`
	TTL    ConfigTTL `json:"ttl"`
	Log    ConfigLog `json:"log"`
}

type ConfigTTL struct {
	Default time.Duration `json:"default"`
	Maximum time.Duration `json:"maximum"`
}

type ConfigLog struct {
	Enabled  bool          `json:"enabled"`
	Channel  string        `json:"channel"`
	Interval time.Duration `json:"interval"`
}

func DefaultConfig() Config {
	return Config{
		Driver: "",
		TTL: ConfigTTL{
			Default: 600 * time.Second,
			Maximum: 0,
		},
		Log: ConfigLog{
			Enabled:  false,
			Channel:  logs.DefaultChannel,
			Interval: 60 * time.Second,
		},
	}
}
