package sqlite3

import (
	"strconv"
	"strings"
)

// Helper implements the SQLite SQL dialect: $N placeholders and
// double-quoted names.
type Helper struct {
}

// ArgPlaceholder returns "$i".
func (h *Helper) ArgPlaceholder(i int) string {
	placeholder := "$" + strconv.Itoa(i)
	return placeholder
}

// EscapeName quotes an identifier with double quotes. See the postgres
// Helper.
func (h *Helper) EscapeName(s string) string {
	s = strings.Replace(s, `"`, "", -1)
	s = strings.Replace(s, ` `, `" "`, -1)
	s = strings.Replace(s, `.`, `"."`, -1)
	s = `"` + s + `"`
	return s
}

// IsDuplicateKey reports whether err is a duplicate key error.
func (h *Helper) IsDuplicateKey(err error) bool {
	return IsDuplicateKey(err)
}
