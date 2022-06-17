package memcached

import (
	"github.com/olegshs/go-tools/cache/storage"
)

type Config struct {
	storage.Config
	Servers []ConfigServer `json:"servers"`
	Prefix  string         `json:"prefix"`
}

type ConfigServer struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

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
