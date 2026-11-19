package session

import (
	"github.com/olegshs/go-tools/events"
)

// Session events, dispatched asynchronously on Manager.Events. The
// comments show the listener arguments.
const (
	EventAfterLoad     = events.Event("AfterLoad")     // (*Session)
	EventBeforeSave    = events.Event("BeforeSave")    // (*Session)
	EventBeforeDestroy = events.Event("BeforeDestroy") // (id string)
)
