package file

// Config holds the settings of the file driver, read from the channel's
// config section.
//
// Path is a file path relative to config.BaseDir, or one of Stdout, Stderr
// or Null. Permissions is the octal mode for new files. Milliseconds adds
// milliseconds to timestamps. Color enables ANSI colors, which is useful
// for terminals.
type Config struct {
	Path         string `json:"path"`
	Permissions  string `json:"permissions"`
	Milliseconds bool   `json:"milliseconds"`
	Color        bool   `json:"color"`
}

// DefaultConfig returns settings that log to stdout without colors.
func DefaultConfig() Config {
	return Config{
		Path:         Stdout,
		Permissions:  "0640",
		Milliseconds: false,
		Color:        false,
	}
}
