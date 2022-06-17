package redis

import (
	"time"

	"github.com/olegshs/go-tools/cache/storage"
)

type Config struct {
	storage.Config
	Host     string     `json:"host"`
	Port     int        `json:"port"`
	DB       int        `json:"db"`
	Password string     `json:"password"`
	Prefix   string     `json:"prefix"`
	Pool     ConfigPool `json:"pool"`
}

type ConfigPool struct {
	MaxIdle         int           `json:"max_idle"`
	MaxActive       int           `json:"max_active"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	Wait            bool          `json:"wait"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime"`
}

func DefaultConfig() Config {
	return Config{
		Config:   storage.DefaultConfig(),
		Host:     "localhost",
		Port:     6379,
		DB:       0,
		Password: "",
		Prefix:   "",
		Pool: ConfigPool{
			MaxIdle:         0,
			MaxActive:       0,
			IdleTimeout:     0,
			Wait:            false,
			MaxConnLifetime: 0,
		},
	}
}
