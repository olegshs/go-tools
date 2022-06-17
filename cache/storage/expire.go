package storage

import (
	"time"
)

const (
	DefaultTTL = 0
	MaximumTTL = -1
)

func Expire(ttl, def, max time.Duration) int64 {
	if ttl == DefaultTTL {
		ttl = def
	}

	if max > 0 && (ttl > max || ttl == MaximumTTL) {
		ttl = max
	}

	var e int64
	if ttl > 0 {
		e = time.Now().Add(ttl).Unix()
	} else {
		e = -1
	}

	return e
}
