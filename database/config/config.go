package config

import (
	"github.com/olegshs/go-tools/logs"
)

type Config struct {
	Driver string `json:"driver"`
	Log    Log    `json:"log"`
}

type Log struct {
	Enabled bool   `json:"enabled"`
	Channel string `json:"channel"`
}

func DefaultConfig() Config {
	return Config{
		Log: Log{
			Enabled: false,
			Channel: logs.DefaultChannel,
		},
	}
}
