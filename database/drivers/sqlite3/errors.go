package sqlite3

import (
	"strings"
)

func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	if strings.Index(err.Error(), "UNIQUE constraint failed:") == 0 {
		return true
	}
	return false
}
