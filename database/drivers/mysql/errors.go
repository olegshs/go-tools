package mysql

import (
	"strings"
)

// IsDuplicateKey reports whether err is a duplicate key error (MySQL error
// 1062).
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	if strings.Index(err.Error(), "Error 1062:") == 0 {
		return true
	}
	return false
}
