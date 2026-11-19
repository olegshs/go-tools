package mysql

import (
	"strings"
)

// Helper implements the MySQL SQL dialect: "?" placeholders and backquoted
// names.
type Helper struct {
}

// ArgPlaceholder returns "?".
func (h *Helper) ArgPlaceholder(int) string {
	return "?"
}

// EscapeName quotes an identifier with backquotes. Dots separate quoted
// parts, so "t.id" becomes `t`.`id`, and a space separates a name from its
// alias. Backquotes in s are removed.
func (h *Helper) EscapeName(s string) string {
	s = strings.Replace(s, "`", "", -1)
	s = strings.Replace(s, " ", "` `", -1)
	s = strings.Replace(s, ".", "`.`", -1)
	s = "`" + s + "`"
	return s
}

// IsDuplicateKey reports whether err is a duplicate key error.
func (h *Helper) IsDuplicateKey(err error) bool {
	return IsDuplicateKey(err)
}
