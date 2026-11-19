// Package logs provides leveled logging through named channels.
//
// A channel is configured under the "logs.<name>" config key. Channels
// without their own section use "logs.default". Each channel has a driver
// (currently only "file"), a minimum level and an optional list of channels
// to forward a copy of every message to:
//
//	logs:
//	  default:
//	    path: /dev/stdout
//	    level: info
//	  db:
//	    path: logs/db.log
//	    forward: [default]
//
// Levels follow RFC 5424 (see package logger). The package-level functions
// such as Info and Error write to the default channel.
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
	// DefaultChannel is the name of the channel used by the package-level
	// functions and the fallback config section for unconfigured channels.
	DefaultChannel = "default"

	// DriverFile writes to a file or a standard stream. It is the default
	// driver.
	DriverFile = "file"
)

var (
	// instances caches channels by name.
	instances      = map[string]*LogChannel{}
	instancesMutex = sync.Mutex{}
)

// LogChannel is a named log destination with its own level filter.
// It is safe for concurrent use.
type LogChannel struct {
	name  string
	log   logger.Logger
	level int
	fwd   []string
}

// Channel returns the channel with the given name, creating it from config on
// first use. It panics if the configured driver is unknown or the log file
// cannot be opened.
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

// getLoggerConfig fills conf from "logs.<name>", or from "logs.default" if
// that section does not exist, and reports whether the named section exists.
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

// Write implements io.Writer. It logs b at info level, without applying the
// level filter or forwarding, so a channel can be passed to log.New or
// similar APIs.
func (c *LogChannel) Write(b []byte) (int, error) {
	return c.log.Write(b)
}

// Print logs the arguments, joined with spaces, at the given level and
// forwards them to the channels listed in the "forward" setting.
// Messages above the channel's level are dropped. Write errors are reported
// on stderr.
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

// forward passes the message on to the forward targets of c. stack holds
// the channels the message has already passed through, which prevents
// infinite loops when channels forward to each other.
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

// Debug logs a message at debug level.
func (c *LogChannel) Debug(a ...interface{}) {
	c.Print(logger.LevelDebug, a...)
}

// Info logs a message at info level.
func (c *LogChannel) Info(a ...interface{}) {
	c.Print(logger.LevelInfo, a...)
}

// Notice logs a message at notice level.
func (c *LogChannel) Notice(a ...interface{}) {
	c.Print(logger.LevelNotice, a...)
}

// Warning logs a message at warning level.
func (c *LogChannel) Warning(a ...interface{}) {
	c.Print(logger.LevelWarning, a...)
}

// Error logs a message at error level.
func (c *LogChannel) Error(a ...interface{}) {
	c.Print(logger.LevelError, a...)
}

// Critical logs a message at critical level.
func (c *LogChannel) Critical(a ...interface{}) {
	c.Print(logger.LevelCritical, a...)
}

// Alert logs a message at alert level.
func (c *LogChannel) Alert(a ...interface{}) {
	c.Print(logger.LevelAlert, a...)
}

// Emergency logs a message at emergency level.
func (c *LogChannel) Emergency(a ...interface{}) {
	c.Print(logger.LevelEmergency, a...)
}
