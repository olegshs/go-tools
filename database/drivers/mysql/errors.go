package mysql

import (
	"strings"
)

func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	if strings.Index(err.Error(), "Error 1062:") == 0 {
		return true
	}
	return false
}
