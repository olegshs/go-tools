package logger

import (
	"fmt"
	"strconv"
	"strings"
)

// Уровни важности сообщений из RFC 5424 (https://tools.ietf.org/html/rfc5424).
const (
	LevelEmergency = iota // system is unusable
	LevelAlert            // action must be taken immediately
	LevelCritical         // critical conditions
	LevelError            // error conditions
	LevelWarning          // warning conditions
	LevelNotice           // normal but significant condition
	LevelInfo             // informational messages
	LevelDebug            // debug-level messages
)

// Названия уровней важности.
const (
	LevelNameEmergency = "emergency"
	LevelNameAlert     = "alert"
	LevelNameCritical  = "critical"
	LevelNameError     = "error"
	LevelNameWarning   = "warning"
	LevelNameNotice    = "notice"
	LevelNameInfo      = "info"
	LevelNameDebug     = "debug"
)

var levelNames = map[int]string{
	LevelEmergency: LevelNameEmergency,
	LevelAlert:     LevelNameAlert,
	LevelCritical:  LevelNameCritical,
	LevelError:     LevelNameError,
	LevelWarning:   LevelNameWarning,
	LevelNotice:    LevelNameNotice,
	LevelInfo:      LevelNameInfo,
	LevelDebug:     LevelNameDebug,
}

var levelByName = map[string]int{}

func init() {
	for level, name := range levelNames {
		levelByName[name] = level
	}
}

// ParseLevel преобразует строку, содержащую номер или название уровня важности, в целое число.
func ParseLevel(s string) int {
	n, err := strconv.ParseInt(s, 0, 32)
	if err == nil {
		return int(n)
	}

	s = strings.ToLower(s)
	level, ok := levelByName[s]
	if !ok {
		level = LevelDebug
	}

	return level
}

// LevelName возвращает название уровня важности по его номеру.
func LevelName(level int) string {
	name, ok := levelNames[level]
	if !ok {
		return fmt.Sprintf("level %d", level)
	}

	return name
}
