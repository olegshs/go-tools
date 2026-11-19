package storage

import (
	"errors"
)

var (
	// Errors returned by the drivers. ErrInvalidData is returned when the data
	// cannot be stored, such as an item larger than the memory cache.
	ErrNotFound    = errors.New("not found")
	ErrExpired     = errors.New("expired")
	ErrInvalidData = errors.New("invalid data")
)
