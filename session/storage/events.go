// Package storage implements session storages that keep encoded session
// data in files or in an SQL table.
package storage

import (
	"github.com/olegshs/go-tools/events"
)

// Storage events, dispatched asynchronously on the storage's Events. The
// comments show the listener arguments.
const (
	EventAfterLoad    = events.Event("AfterLoad")    // (id string, data []byte)
	EventBeforeSave   = events.Event("BeforeSave")   // (id string, data []byte)
	EventBeforeTouch  = events.Event("BeforeTouch")  // (id string)
	EventBeforeDelete = events.Event("BeforeDelete") // (id string)
)
