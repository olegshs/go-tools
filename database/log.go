package database

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/olegshs/go-tools/database/config"
	"github.com/olegshs/go-tools/database/interfaces"
	"github.com/olegshs/go-tools/helpers/typeconv"
	"github.com/olegshs/go-tools/logs"
)

const (
	// maxArgLength is the length at which logged argument values are cut off.
	maxArgLength = 100
)

// Log writes every query of a connection to a log channel: successful
// queries at debug level and failed ones at error level, with their
// arguments and duration. Byte slice arguments are logged in base64.
type Log struct {
	db      interfaces.DB
	channel *logs.LogChannel
	count   atomic.Int64

	// listener is log.onEvent, stored once because every evaluation of a
	// method value creates a new function that RemoveListener cannot match.
	listener func(time.Time, time.Time, string, []interface{}, error)
}

// NewLog creates a query log for db. Call Start to begin.
func NewLog(db interfaces.DB, conf config.Log) *Log {
	log := new(Log)
	log.db = db
	log.channel = logs.Channel(conf.Channel)
	log.listener = log.onEvent

	return log
}

// Start begins logging by listening to the query events of db.
func (log *Log) Start() {
	log.db.Events().AddListener(EventPrepare, log.listener)
	log.db.Events().AddListener(EventExec, log.listener)
	log.db.Events().AddListener(EventQuery, log.listener)
	log.db.Events().AddListener(EventQueryRow, log.listener)
}

// Stop stops logging.
func (log *Log) Stop() {
	log.db.Events().RemoveListener(EventPrepare, log.listener)
	log.db.Events().RemoveListener(EventExec, log.listener)
	log.db.Events().RemoveListener(EventQuery, log.listener)
	log.db.Events().RemoveListener(EventQueryRow, log.listener)
}

// onEvent logs one query with a sequence number.
func (log *Log) onEvent(startTime time.Time, endTime time.Time, query string, args []interface{}, err error) {
	n := log.count.Add(1)

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
			n, query, argsStr, durationStr,
		))
	} else {
		log.channel.Error("database:", fmt.Sprintf(
			"query %d:\n%s\n%s%s\n%s",
			n, query, argsStr, durationStr, err.Error(),
		))
	}
}

// formatDuration formats d in milliseconds with about three significant
// digits.
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
