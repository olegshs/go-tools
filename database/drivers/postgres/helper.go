package postgres

import (
	"strconv"
	"strings"
)

// Helper implements the PostgreSQL SQL dialect: $N placeholders and
// double-quoted names.
type Helper struct {
}

// ArgPlaceholder returns "$i".
func (h *Helper) ArgPlaceholder(i int) string {
	placeholder := "$" + strconv.Itoa(i)
	return placeholder
}

// EscapeName quotes an identifier with double quotes. Dots separate quoted
// parts, so "t.id" becomes "t"."id", and a space separates a name from its
// alias. Double quotes in s are removed.
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
