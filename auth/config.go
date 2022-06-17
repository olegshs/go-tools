package auth

import (
	"time"
)

type Config struct {
	Password ConfigPassword `json:"password"`
	Cookie   ConfigCookie   `json:"cookie"`
	Session  ConfigSession  `json:"session"`
}

type ConfigPassword struct {
	Hash string `json:"hash"`
	Salt string `json:"salt"`
}

type ConfigCookie struct {
	Name   string        `json:"name"`
	Path   string        `json:"path"`
	Domain string        `json:"domain"`
	Secure bool          `json:"secure"`
	TTL    time.Duration `json:"ttl"`
	Hash   string        `json:"hash"`
	Secret string        `json:"secret"`
}

type ConfigSession struct {
	UserIdKey string `json:"user_id_key"`
}

func DefaultConfig() Config {
	return Config{
		Password: ConfigPassword{
			Hash: "sha256",
			Salt: "",
		},
		Cookie: ConfigCookie{
			Name:   "auth",
			Path:   "/",
			Domain: "",
			Secure: false,
			TTL:    7 * 24 * time.Hour,
			Hash:   "sha256",
			Secret: "",
		},
		Session: ConfigSession{
			UserIdKey: "auth.user.id",
		},
	}
}
