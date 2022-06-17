package file

type Config struct {
	Path         string `json:"path"`
	Permissions  string `json:"permissions"`
	Milliseconds bool   `json:"milliseconds"`
	Color        bool   `json:"color"`
}

func DefaultConfig() Config {
	return Config{
		Path:         Stdout,
		Permissions:  "0640",
		Milliseconds: false,
		Color:        false,
	}
}
