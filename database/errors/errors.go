// Package errors defines the errors shared by the database packages.
package errors

import (
	"errors"
)

var (
	// UnknownDriver is returned when the configured driver is not supported.
	UnknownDriver = errors.New("unknown driver")
	// UnknownDatabase is returned when a connection name has no config section.
	UnknownDatabase = errors.New("unknown database")
	// InvalidStatement is returned for a query that cannot be built.
	InvalidStatement = errors.New("invalid statement")
)
