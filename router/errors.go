package router

import (
	"errors"
)

var (
	// Errors returned by URL generation.
	ErrRouteNotFound       = errors.New("route not found")
	ErrNotEnoughParameters = errors.New("not enough parameters")
	ErrInvalidParameter    = errors.New("invalid parameter")
)
