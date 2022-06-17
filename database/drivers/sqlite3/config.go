package sqlite3

import (
	"github.com/olegshs/go-tools/database/config"
)

type Config struct {
	config.Config
	File   string        `json:"file"`
	Params config.Params `json:"params"`
}

func DefaultConfig() Config {
	return Config{
		Config: config.DefaultConfig(),
		File:   "test.db",
		Params: config.Params{},
	}
}
