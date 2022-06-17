package session

import (
	"time"
)

type Config struct {
	Id      ConfigId      `json:"id"`
	Cookie  ConfigCookie  `json:"cookie"`
	TTL     time.Duration `json:"ttl"`
	Storage ConfigStorage `json:"storage"`
}

type ConfigId struct {
	Length int `json:"length"`
}

type ConfigCookie struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Domain string `json:"domain"`
	Secure bool   `json:"secure"`
}

type ConfigStorage struct {
	Driver string `json:"driver"`
}

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
