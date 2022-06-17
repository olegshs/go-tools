package server

import (
	"time"

	"github.com/olegshs/go-tools/logs"
)

type Config struct {
	Listen    string          `json:"listen"`
	Static    ConfigStatic    `json:"static"`
	Log       ConfigLog       `json:"log"`
	AccessLog ConfigAccessLog `json:"access_log"`
	Tokens    bool            `json:"tokens"`

	ReadTimeout       time.Duration `json:"read_timeout"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	MaxHeaderBytes    int           `json:"max_header_bytes"`
}

type ConfigStatic struct {
	Enabled bool   `json:"enabled"`
	Dir     string `json:"dir"`
}

type ConfigLog struct {
	Enabled bool   `json:"enabled"`
	Channel string `json:"channel"`
}

type ConfigAccessLog struct {
	Enabled bool   `json:"enabled"`
	Channel string `json:"channel"`
}

func DefaultConfig() Config {
	return Config{
		Listen: ":8080",
		Static: ConfigStatic{
			Enabled: true,
			Dir:     "public",
		},
		Log: ConfigLog{
			Enabled: true,
			Channel: logs.DefaultChannel,
		},
		AccessLog: ConfigAccessLog{
			Enabled: false,
			Channel: logs.DefaultChannel,
		},
		Tokens:            true,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
	}
}
