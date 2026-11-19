// Package config holds the connection settings shared by all database
// drivers.
package config

import (
	"github.com/olegshs/go-tools/logs"
)

// Config holds the driver-independent connection settings. Drivers embed
// it in their own Config.
type Config struct {
	Driver string `json:"driver"`
	Log    Log    `json:"log"`
}

// Log enables the query log on a log channel.
type Log struct {
	Enabled bool   `json:"enabled"`
	Channel string `json:"channel"`
}

// DefaultConfig returns the default settings, with the query log disabled.
func DefaultConfig() Config {
	return Config{
		Log: Log{
			Enabled: false,
			Channel: logs.DefaultChannel,
		},
	}
}
