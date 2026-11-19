package helpers

import (
	"strings"
)

const (
	// trimCutset lists the characters removed by Trim, TrimLeft and TrimRight:
	// ASCII whitespace plus NUL.
	trimCutset = " \t\r\n\x00\v"
)

// Trim removes leading and trailing whitespace and NUL characters from s.
func Trim(s string) string {
	return strings.Trim(s, trimCutset)
}

// TrimLeft removes leading whitespace and NUL characters from s.
func TrimLeft(s string) string {
	return strings.TrimLeft(s, trimCutset)
}

// TrimRight removes trailing whitespace and NUL characters from s.
func TrimRight(s string) string {
	return strings.TrimRight(s, trimCutset)
}
