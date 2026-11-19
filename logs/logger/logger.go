// Package logger holds the definitions shared by the log drivers: the Logger
// interface, the channel config and the RFC 5424 severity levels.
package logger

import (
	"io"
)

// Logger is the interface that log drivers implement.
//
// Write writes raw bytes as an info-level line. Print formats the arguments
// and writes them as one line at the given level. Level filtering is done by
// the channel, not by the driver.
type Logger interface {
	io.Writer
	Print(level int, a ...interface{}) error
}
