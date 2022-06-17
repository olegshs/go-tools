package database

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/olegshs/go-tools/database/config"
	"github.com/olegshs/go-tools/database/interfaces"
	"github.com/olegshs/go-tools/helpers/typeconv"
	"github.com/olegshs/go-tools/logs"
)

const (
	maxArgLength = 100
)

type Log struct {
	db      interfaces.DB
	channel *logs.LogChannel
	count   int
}

func NewLog(db interfaces.DB, conf config.Log) *Log {
	log := new(Log)
	log.db = db
	log.channel = logs.Channel(conf.Channel)
	log.count = 0

	return log
}

func (log *Log) Start() {
	log.db.Events().AddListener(EventPrepare, log.onEvent)
	log.db.Events().AddListener(EventExec, log.onEvent)
	log.db.Events().AddListener(EventQuery, log.onEvent)
	log.db.Events().AddListener(EventQueryRow, log.onEvent)
}

func (log *Log) Stop() {
	log.db.Events().RemoveListener(EventPrepare, log.onEvent)
	log.db.Events().RemoveListener(EventExec, log.onEvent)
	log.db.Events().RemoveListener(EventQuery, log.onEvent)
	log.db.Events().RemoveListener(EventQueryRow, log.onEvent)
}

func (log *Log) onEvent(startTime time.Time, endTime time.Time, query string, args []interface{}, err error) {
	log.count++

	argsStr := ""
	if args != nil {
		if len(args) > 0 {
			a := make([]string, len(args))

			for i, v := range args {
				var s string

				switch t := v.(type) {
				case []byte:
					if len(t) > maxArgLength {
						t = t[:maxArgLength]
					}
					s = base64.StdEncoding.EncodeToString(t)
				default:
					s = typeconv.String(v)
				}

				if len(s) > maxArgLength {
					s = s[:maxArgLength] + "…"
				}

				a[i] = s
			}

			argsStr = strings.Join(a, ", ")
			argsStr = "[ " + argsStr + " ]\n"
		}
	} else {
		argsStr = "(prepare)\n"
	}

	duration := endTime.Sub(startTime)
	durationStr := log.formatDuration(duration)

	if err == nil {
		log.channel.Debug("database:", fmt.Sprintf(
			"query %d:\n%s\n%s%s",
			log.count, query, argsStr, durationStr,
		))
	} else {
		log.channel.Error("database:", fmt.Sprintf(
			"query %d:\n%s\n%s%s\n%s",
			log.count, query, argsStr, durationStr, err.Error(),
		))
	}
}

func (log *Log) formatDuration(d time.Duration) string {
	ms := float64(d) / float64(time.Millisecond)

	var format string
	if ms < 10 {
		format = "%.3f"
	} else if ms < 100 {
		format = "%.1f"
	} else {
		format = "%.0f"
	}

	return fmt.Sprintf(format+" ms", ms)
}
