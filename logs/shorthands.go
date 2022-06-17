package logs

// Write добавляет сообщение в журнал.
func Write(b []byte) (int, error) {
	return Channel(DefaultChannel).Write(b)
}

// Print добавляет сообщение с заданным уровнем важности в журнал.
func Print(level int, a ...interface{}) {
	Channel(DefaultChannel).Print(level, a...)
}

// Debug добавляет сообщение с уровнем важности debug в журнал.
func Debug(a ...interface{}) {
	Channel(DefaultChannel).Debug(a...)
}

// Info добавляет сообщение с уровнем важности info в журнал.
func Info(a ...interface{}) {
	Channel(DefaultChannel).Info(a...)
}

// Notice добавляет сообщение с уровнем важности notice в журнал.
func Notice(a ...interface{}) {
	Channel(DefaultChannel).Notice(a...)
}

// Warning добавляет сообщение с уровнем важности warning в журнал.
func Warning(a ...interface{}) {
	Channel(DefaultChannel).Warning(a...)
}

// Error добавляет сообщение с уровнем важности error в журнал.
func Error(a ...interface{}) {
	Channel(DefaultChannel).Error(a...)
}

// Critical добавляет сообщение с уровнем важности critical в журнал.
func Critical(a ...interface{}) {
	Channel(DefaultChannel).Critical(a...)
}

// Alert добавляет сообщение с уровнем важности alert в журнал.
func Alert(a ...interface{}) {
	Channel(DefaultChannel).Alert(a...)
}

// Emergency добавляет сообщение с уровнем важности emergency в журнал.
func Emergency(a ...interface{}) {
	Channel(DefaultChannel).Emergency(a...)
}
