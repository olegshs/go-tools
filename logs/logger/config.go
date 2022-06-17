package logger

// Конфигурация канала.
type Config struct {
	// Название драйвера.
	Driver string `json:"driver"`

	// Минимальный уровень важности сообщений, добавляемых в журнал.
	Level string `json:"level"`

	// Список каналов, в которые будут отправляться копии сообщений.
	Forward []string `json:"forward"`
}

// DefaultConfig возвращает конфигурацию по умолчанию.
func DefaultConfig() Config {
	return Config{
		Driver:  "",
		Level:   LevelNameDebug,
		Forward: nil,
	}
}
