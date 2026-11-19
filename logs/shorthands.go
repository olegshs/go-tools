package logs

// Write writes b to the default channel. See LogChannel.Write.
func Write(b []byte) (int, error) {
	return Channel(DefaultChannel).Write(b)
}

// Print logs a message at the given level to the default channel.
func Print(level int, a ...interface{}) {
	Channel(DefaultChannel).Print(level, a...)
}

// Debug logs a message at debug level to the default channel.
func Debug(a ...interface{}) {
	Channel(DefaultChannel).Debug(a...)
}

// Info logs a message at info level to the default channel.
func Info(a ...interface{}) {
	Channel(DefaultChannel).Info(a...)
}

// Notice logs a message at notice level to the default channel.
func Notice(a ...interface{}) {
	Channel(DefaultChannel).Notice(a...)
}

// Warning logs a message at warning level to the default channel.
func Warning(a ...interface{}) {
	Channel(DefaultChannel).Warning(a...)
}

// Error logs a message at error level to the default channel.
func Error(a ...interface{}) {
	Channel(DefaultChannel).Error(a...)
}

// Critical logs a message at critical level to the default channel.
func Critical(a ...interface{}) {
	Channel(DefaultChannel).Critical(a...)
}

// Alert logs a message at alert level to the default channel.
func Alert(a ...interface{}) {
	Channel(DefaultChannel).Alert(a...)
}

// Emergency logs a message at emergency level to the default channel.
func Emergency(a ...interface{}) {
	Channel(DefaultChannel).Emergency(a...)
}
