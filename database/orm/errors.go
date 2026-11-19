package orm

import (
	"database/sql"
	"errors"
)

var (
	// Errors returned by queries. ErrNoRows is sql.ErrNoRows, so either can be
	// compared with.
	ErrInvalidModel = errors.New("invalid model")
	ErrNotPointer   = errors.New("output is not a pointer")
	ErrNotSlice     = errors.New("output is not a pointer to slice")
	ErrNoPrimaryKey = errors.New("primary key is not defined or empty")
	ErrNoRows       = sql.ErrNoRows
)

// ErrDuplicateKey wraps the driver error of a unique constraint violation.
type ErrDuplicateKey struct {
	err error
}

// Error returns the message of the driver error.
func (e ErrDuplicateKey) Error() string {
	return e.err.Error()
}

// IsDuplicateKey reports whether err is an ErrDuplicateKey.
func IsDuplicateKey(err error) bool {
	_, ok := err.(ErrDuplicateKey)
	return ok
}
