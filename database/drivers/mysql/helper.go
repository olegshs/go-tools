package mysql

import (
	"strings"
)

type Helper struct {
}

func (h *Helper) ArgPlaceholder(int) string {
	return "?"
}

func (h *Helper) EscapeName(s string) string {
	s = strings.Replace(s, "`", "", -1)
	s = strings.Replace(s, " ", "` `", -1)
	s = strings.Replace(s, ".", "`.`", -1)
	s = "`" + s + "`"
	return s
}

func (h *Helper) IsDuplicateKey(err error) bool {
	return IsDuplicateKey(err)
}
