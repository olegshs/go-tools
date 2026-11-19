package logger

// Config holds the driver-independent settings of a log channel.
type Config struct {
	// Driver is the driver name. Empty means "file".
	Driver string `json:"driver"`

	// Level is the least severe level that is logged, as a name ("warning") or
	// a number. Defaults to "debug", which logs everything.
	Level string `json:"level"`

	// Forward lists channels that receive a copy of every message logged to
	// this channel, filtered by their own levels.
	Forward []string `json:"forward"`
}

// DefaultConfig returns the default channel settings.
func DefaultConfig() Config {
	return Config{
		Driver:  "",
		Level:   LevelNameDebug,
		Forward: nil,
	}
}
