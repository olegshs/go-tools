package storage

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExpired  = errors.New("expired")
)
