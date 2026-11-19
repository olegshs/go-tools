package memcached

import (
	"github.com/olegshs/go-tools/cache/storage"
)

// Config holds the Memcached driver settings. Prefix is added to every key,
// after the "(name)" prefix of the storage.
type Config struct {
	storage.Config
	Servers []ConfigServer `json:"servers"`
	Prefix  string         `json:"prefix"`
}

// ConfigServer is the address of a Memcached server.
type ConfigServer struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// DefaultConfig returns settings for a single server on localhost:11211.
func DefaultConfig() Config {
	return Config{
		Config: storage.DefaultConfig(),
		Servers: []ConfigServer{
			{
				Host: "localhost",
				Port: 11211,
			},
		},
		Prefix: "",
	}
}
