package orm

import (
	"database/sql"
	"errors"
)

var (
	ErrInvalidModel = errors.New("invalid model")
	ErrNotPointer   = errors.New("output is not a pointer")
	ErrNotSlice     = errors.New("output is not a pointer to slice")
	ErrNoPrimaryKey = errors.New("primary key is not defined or empty")
	ErrNoRows       = sql.ErrNoRows
)

type ErrDuplicateKey struct {
	err error
}

func (e ErrDuplicateKey) Error() string {
	return e.err.Error()
}

func IsDuplicateKey(err error) bool {
	_, ok := err.(ErrDuplicateKey)
	return ok
}
