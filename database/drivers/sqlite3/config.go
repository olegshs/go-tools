package sqlite3

import (
	"github.com/olegshs/go-tools/database/config"
)

// Config holds the SQLite settings. File is the database file path, and
// Params are appended to the DSN as URI parameters.
type Config struct {
	config.Config
	File   string        `json:"file"`
	Params config.Params `json:"params"`
}

// DefaultConfig returns settings for "test.db" in the working directory.
func DefaultConfig() Config {
	return Config{
		Config: config.DefaultConfig(),
		File:   "test.db",
		Params: config.Params{},
	}
}
