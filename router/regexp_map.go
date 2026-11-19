package router

import (
	"regexp"
)

// regexpMap caches compiled regexps by source string. It is not safe for
// concurrent use.
type regexpMap map[string]*regexp.Regexp

// Get returns the compiled regexp for s, compiling it on first use. It panics
// if s is invalid.
func (m regexpMap) Get(s string) *regexp.Regexp {
	r, ok := m[s]
	if !ok {
		r = regexp.MustCompile(s)
		m[s] = r
	}
	return r
}
