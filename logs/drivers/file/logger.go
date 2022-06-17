// Пакет file реализует драйвер для записи журналов в файлы.
package file

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olegshs/go-tools/events"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/helpers/typeconv"
	"github.com/olegshs/go-tools/logs/logger"
)

const (
	Stdout = "/dev/stdout"
	Stderr = "/dev/stderr"
	Null   = "/dev/null"

	TimeFormatDefault      = "2006-01-02 15:04:05"
	TimeFormatMilliseconds = "2006-01-02 15:04:05.000"
)

var levelColors = map[int]int{
	logger.LevelEmergency: 31,
	logger.LevelAlert:     31,
	logger.LevelCritical:  31,
	logger.LevelError:     31,
	logger.LevelWarning:   33,
	logger.LevelNotice:    32,
	logger.LevelInfo:      0,
	logger.LevelDebug:     2,
}

type Logger struct {
	out        io.Writer
	timeFormat string
	color      bool
}

func New(conf Config) *Logger {
	log := new(Logger)

	switch conf.Path {
	case Stdout:
		log.out = os.Stdout
	case Stderr:
		log.out = os.Stderr
	case Null:
		log.out = ioutil.Discard
	default:
		err := log.setFile(conf.Path, conf.Permissions)
		if err != nil {
			panic(err)
		}
	}

	if conf.Milliseconds {
		log.timeFormat = TimeFormatMilliseconds
	} else {
		log.timeFormat = TimeFormatDefault
	}

	log.color = conf.Color

	events.Shared.AddListener(events.EventExit, func() {
		err := log.Flush()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	})

	return log
}

func (log *Logger) Write(b []byte) (int, error) {
	s := log.format(logger.LevelInfo, string(b))

	_, err := log.out.Write([]byte(s))
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (log *Logger) Print(level int, a ...interface{}) error {
	s := log.format(level, a...)

	_, err := log.out.Write([]byte(s))
	if err != nil {
		return err
	}

	return nil
}

func (log *Logger) Flush() error {
	switch t := log.out.(type) {
	case interface{ Flush() error }:
		return t.Flush()
	default:
		return nil
	}
}

func (log *Logger) setFile(filename string, permissions string) error {
	perm, err := strconv.ParseInt(permissions, 8, 10)
	if err != nil {
		return fmt.Errorf("invalid permissions: %w", err)
	}

	w, err := openFile(filename, os.FileMode(perm))
	if err != nil {
		return err
	}

	log.out = w
	return nil
}

func (log *Logger) format(level int, a ...interface{}) string {
	ts := log.time()

	n := len(a)
	s := make([]string, n)
	for i := 0; i < n; i++ {
		s[i] = typeconv.String(a[i])
	}

	line := log.colorLevel(level, log.colorBold(ts)) + " " +
		log.colorLevel(level, logger.LevelName(level)+":") + " " +
		log.colorLevel(level, helpers.TrimRight(strings.Join(s, " "))) +
		"\n"

	return line
}

func (log *Logger) time() string {
	return time.Now().UTC().Format(log.timeFormat)
}

func (log *Logger) colorBold(s string) string {
	if !log.color {
		return s
	}
	return log.ansiColor(1, s)
}

func (log *Logger) colorLevel(level int, s string) string {
	if !log.color {
		return s
	}
	return log.ansiColor(levelColors[level], s)
}

func (log *Logger) ansiColor(color int, s string) string {
	if color <= 0 {
		return s
	}
	return fmt.Sprintf("\x1B[%dm%s\x1B[m", color, s)
}
