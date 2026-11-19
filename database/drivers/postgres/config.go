package postgres

import (
	"github.com/olegshs/go-tools/database/config"
)

// Config holds the PostgreSQL connection settings. Params are appended to
// the connection URL.
type Config struct {
	config.Config
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	Database string        `json:"database"`
	Params   config.Params `json:"params"`
}

// DefaultConfig returns settings for a local server with SSL disabled.
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
