package database

import (
	"github.com/olegshs/go-tools/events"
)

// Query events, dispatched asynchronously on DB.Events. Listeners receive
// the arguments shown in the comments: the start and end time of the call,
// the SQL, its arguments and the error, if any.
const (
	EventPing     = events.Event("Ping")     // (startTime, endTime, nil, nil, err)
	EventExec     = events.Event("Exec")     // (startTime, endTime, query, args, err)
	EventPrepare  = events.Event("Prepare")  // (startTime, endTime, query, nil, err)
	EventQuery    = events.Event("Query")    // (startTime, endTime, query, args, err)
	EventQueryRow = events.Event("QueryRow") // (startTime, endTime, query, args, err)
)
