package session

import (
	"time"
)

// Config holds the session settings, read from the "session" config section.
// TTL is how long an unused session is kept. Zero means sessions never
// expire.
type Config struct {
	Id      ConfigId      `json:"id"`
	Cookie  ConfigCookie  `json:"cookie"`
	TTL     time.Duration `json:"ttl"`
	Storage ConfigStorage `json:"storage"`
}

// ConfigId sets the session ID length in random bytes. The ID string is
// twice as long.
type ConfigId struct {
	Length int `json:"length"`
}

// ConfigCookie holds the session cookie attributes.
type ConfigCookie struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Domain string `json:"domain"`
	Secure bool   `json:"secure"`
}

// ConfigStorage selects the storage driver: "file" or "sql". The driver's
// own settings are read from the same "session.storage" section.
type ConfigStorage struct {
	Driver string `json:"driver"`
}

// DefaultConfig returns the default settings: 16-byte IDs in a "sess_id"
// cookie, file storage and no expiry.
func DefaultConfig() Config {
	return Config{
		Id: ConfigId{
			Length: 16,
		},
		Cookie: ConfigCookie{
			Name:   "sess_id",
			Path:   "/",
			Domain: "",
			Secure: false,
		},
		Storage: ConfigStorage{
			Driver: "file",
		},
	}
}
