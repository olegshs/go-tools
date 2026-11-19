package storage

import (
	"time"
)

const (
	// Special TTL values. See cache.DefaultTTL and cache.MaximumTTL.
	DefaultTTL = 0
	MaximumTTL = -1
)

// Expire returns the expiry time of an entry stored now with ttl, as a Unix
// timestamp, or -1 if it never expires. def replaces DefaultTTL, and a
// positive max caps ttl and replaces MaximumTTL.
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
