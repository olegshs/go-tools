package helpers

import (
	"strings"
)

const (
	trimCutset = " \t\r\n\x00\v"
)

func Trim(s string) string {
	return strings.Trim(s, trimCutset)
}

func TrimLeft(s string) string {
	return strings.TrimLeft(s, trimCutset)
}

func TrimRight(s string) string {
	return strings.TrimRight(s, trimCutset)
}
