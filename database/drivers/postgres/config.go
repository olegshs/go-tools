package postgres

import (
	"github.com/olegshs/go-tools/database/config"
)

type Config struct {
	config.Config
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	Database string        `json:"database"`
	Params   config.Params `json:"params"`
}

func DefaultConfig() Config {
	return Config{
		Config:   config.DefaultConfig(),
		Host:     "localhost",
		Port:     5432,
		Username: "test",
		Password: "test",
		Database: "test",
		Params: config.Params{
			"sslmode": "disable",
		},
	}
}
