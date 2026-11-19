package postgres

import (
	"strings"
)

// IsDuplicateKey reports whether err is a unique violation reported by
// github.com/lib/pq.
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	if strings.Index(err.Error(), "pq: duplicate key") == 0 {
		return true
	}
	return false
}
