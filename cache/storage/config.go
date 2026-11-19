// Package storage holds the definitions shared by the cache drivers:
// common config, errors and TTL handling.
package storage

import (
	"time"

	"github.com/olegshs/go-tools/logs"
)

// Config holds the settings shared by all drivers. Drivers embed it in
// their own Config.
type Config struct {
	Driver string    `json:"driver"`
	TTL    ConfigTTL `json:"ttl"`
	Log    ConfigLog `json:"log"`
}

// ConfigTTL sets the TTL used for DefaultTTL and the maximum TTL. Zero
// Default means no expiry, and zero Maximum means no limit.
type ConfigTTL struct {
	Default time.Duration `json:"default"`
	Maximum time.Duration `json:"maximum"`
}

// ConfigLog enables a periodic log of hit and miss counts.
type ConfigLog struct {
	Enabled  bool          `json:"enabled"`
	Channel  string        `json:"channel"`
	Interval time.Duration `json:"interval"`
}

// DefaultConfig returns the shared defaults: a 10-minute default TTL, no
// maximum, and no statistics log.
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
