package storage

import (
	"errors"
)

var (
	// ErrNotFound is returned when a session does not exist.
	ErrNotFound = errors.New("not found")
	// ErrExpired is returned when a session exists but is older than the TTL.
	ErrExpired = errors.New("expired")
)
