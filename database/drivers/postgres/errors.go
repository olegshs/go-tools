package postgres

import (
	"strings"
)

func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	if strings.Index(err.Error(), "pq: duplicate key") == 0 {
		return true
	}
	return false
}
