package server

import (
	"time"

	"github.com/olegshs/go-tools/logs"
)

// Config holds the server settings, read from the "server" config section.
//
// Listen is the address to listen on, such as ":8080". Tokens adds the
// "X-Powered-By: go-tools" response header. The timeouts and MaxHeaderBytes
// are passed to http.Server, where zero means no limit or the default.
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

// ConfigStatic enables serving files from Dir, relative to config.BaseDir.
type ConfigStatic struct {
	Enabled bool   `json:"enabled"`
	Dir     string `json:"dir"`
}

// ConfigLog enables server status messages, such as the listen address,
// on a log channel.
type ConfigLog struct {
	Enabled bool   `json:"enabled"`
	Channel string `json:"channel"`
}

// ConfigAccessLog enables logging the method and URL of every request on a
// log channel.
type ConfigAccessLog struct {
	Enabled bool   `json:"enabled"`
	Channel string `json:"channel"`
}

// DefaultConfig returns the default settings: listen on :8080 and serve
// static files from "public".
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
