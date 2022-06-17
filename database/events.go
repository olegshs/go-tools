package database

import (
	"github.com/olegshs/go-tools/events"
)

const (
	EventPing     = events.Event("Ping")     // (startTime, endTime, nil, nil, err)
	EventExec     = events.Event("Exec")     // (startTime, endTime, query, args, err)
	EventPrepare  = events.Event("Prepare")  // (startTime, endTime, query, nil, err)
	EventQuery    = events.Event("Query")    // (startTime, endTime, query, args, err)
	EventQueryRow = events.Event("QueryRow") // (startTime, endTime, query, args, err)
)
