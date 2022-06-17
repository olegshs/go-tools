package storage

import (
	"github.com/olegshs/go-tools/events"
)

const (
	EventAfterLoad    = events.Event("AfterLoad")    // (id string, data []byte)
	EventBeforeSave   = events.Event("BeforeSave")   // (id string, data []byte)
	EventBeforeTouch  = events.Event("BeforeTouch")  // (id string)
	EventBeforeDelete = events.Event("BeforeDelete") // (id string)
)
