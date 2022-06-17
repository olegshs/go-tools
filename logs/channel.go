// Пакет logs реализует журнал сообщений и ошибок.
package logs

import (
	"fmt"
	"os"
	"sync"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/logs/drivers/file"
	"github.com/olegshs/go-tools/logs/logger"
)

const (
	// Название канала по умолчанию.
	DefaultChannel = "default"

	// Драйверы:
	DriverFile = "file"
)

var (
	instances      = map[string]*LogChannel{}
	instancesMutex = sync.Mutex{}
)

// Канал.
type LogChannel struct {
	name  string
	log   logger.Logger
	level int
	fwd   []string
}

// Channel возвращает канал по его названию.
func Channel(name string) *LogChannel {
	instancesMutex.Lock()
	defer instancesMutex.Unlock()

	log, ok := instances[name]
	if ok {
		return log
	}

	log = newChannel(name)
	instances[name] = log

	return log
}

func newChannel(name string) *LogChannel {
	conf := logger.DefaultConfig()
	getLoggerConfig(name, &conf)

	c := new(LogChannel)
	c.name = name
	c.log = newLogger(conf.Driver, name)
	c.level = logger.ParseLevel(conf.Level)
	c.fwd = conf.Forward

	return c
}

func newLogger(driver string, name string) logger.Logger {
	switch driver {
	case DriverFile, "":
		conf := file.DefaultConfig()
		getLoggerConfig(name, &conf)
		return file.New(conf)

	default:
		panic("unknown driver: " + driver)
	}
}

func getLoggerConfig(name string, conf interface{}) (exists bool) {
	const prefix = "logs."
	key := prefix + name

	exists = config.Exists(key)
	if !exists {
		key = prefix + DefaultChannel
	}

	config.GetStruct(key, conf)
	return
}

// Write реализует интерфейс io.Writer.
func (c *LogChannel) Write(b []byte) (int, error) {
	return c.log.Write(b)
}

// Print добавляет в журнал сообщение с заданным уровнем важности.
func (c *LogChannel) Print(level int, a ...interface{}) {
	c.print(level, a...)
	c.forward(helpers.Slice[string]{}, level, a...)
}

func (c *LogChannel) print(level int, a ...interface{}) {
	if level > c.level {
		return
	}

	err := c.log.Print(level, a...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func (c *LogChannel) forward(stack helpers.Slice[string], level int, a ...interface{}) {
	if len(c.fwd) == 0 {
		return
	}

	stack = append(stack, c.name)

	for _, name := range c.fwd {
		if stack.IndexOf(name) >= 0 {
			continue
		}

		Channel(name).printForwarded(stack, level, a...)
	}
}

func (c *LogChannel) printForwarded(stack helpers.Slice[string], level int, a ...interface{}) {
	c.print(level, a...)
	c.forward(stack, level, a...)
}

// Debug добавляет в журнал сообщение с уровнем важности debug.
func (c *LogChannel) Debug(a ...interface{}) {
	c.Print(logger.LevelDebug, a...)
}

// Info добавляет в журнал сообщение с уровнем важности info.
func (c *LogChannel) Info(a ...interface{}) {
	c.Print(logger.LevelInfo, a...)
}

// Notice добавляет в журнал сообщение с уровнем важности notice.
func (c *LogChannel) Notice(a ...interface{}) {
	c.Print(logger.LevelNotice, a...)
}

// Warning добавляет в журнал сообщение с уровнем важности warning.
func (c *LogChannel) Warning(a ...interface{}) {
	c.Print(logger.LevelWarning, a...)
}

// Error добавляет в журнал сообщение с уровнем важности error.
func (c *LogChannel) Error(a ...interface{}) {
	c.Print(logger.LevelError, a...)
}

// Critical добавляет в журнал сообщение с уровнем важности critical.
func (c *LogChannel) Critical(a ...interface{}) {
	c.Print(logger.LevelCritical, a...)
}

// Alert добавляет в журнал сообщение с уровнем важности alert.
func (c *LogChannel) Alert(a ...interface{}) {
	c.Print(logger.LevelAlert, a...)
}

// Emergency добавляет в журнал сообщение с уровнем важности emergency.
func (c *LogChannel) Emergency(a ...interface{}) {
	c.Print(logger.LevelEmergency, a...)
}
