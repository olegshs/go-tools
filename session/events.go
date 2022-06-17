package session

import (
	"github.com/olegshs/go-tools/events"
)

const (
	EventAfterLoad     = events.Event("AfterLoad")     // (*Session)
	EventBeforeSave    = events.Event("BeforeSave")    // (*Session)
	EventBeforeDestroy = events.Event("BeforeDestroy") // (id string)
)
