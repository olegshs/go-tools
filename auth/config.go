package auth

import (
	"time"
)

// Config holds the auth settings, read from the "auth" config section.
type Config struct {
	Password ConfigPassword `json:"password"`
	Cookie   ConfigCookie   `json:"cookie"`
	Session  ConfigSession  `json:"session"`
}

// ConfigPassword sets how passwords are hashed. Hash is a name accepted by
// helpers.HashByName, or "none" to store passwords in plain text. Salt is a
// global salt added to every password in addition to the user's own salt.
type ConfigPassword struct {
	Hash string `json:"hash"`
	Salt string `json:"salt"`
}

// ConfigCookie holds the remember-me cookie settings. TTL is how long the
// cookie is valid. Hash names the hash function used to sign it, and Secret
// is the signing key.
type ConfigCookie struct {
	Name   string        `json:"name"`
	Path   string        `json:"path"`
	Domain string        `json:"domain"`
	Secure bool          `json:"secure"`
	TTL    time.Duration `json:"ttl"`
	Hash   string        `json:"hash"`
	Secret string        `json:"secret"`
}

// ConfigSession sets the session key that holds the logged-in user's ID.
type ConfigSession struct {
	UserIdKey string `json:"user_id_key"`
}

// DefaultConfig returns the default settings: SHA-256 password hashes and a
// one-week remember-me cookie. The secret is empty and must be configured.
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
