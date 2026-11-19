package router

import (
	"regexp"

	"github.com/julienschmidt/httprouter"
)

// conditions maps parameter positions to the regexps their values must
// match.
type conditions map[int]*regexp.Regexp

func (c conditions) clone() conditions {
	clone := make(conditions, len(c))
	for k, r := range c {
		clone[k] = r
	}
	return clone
}

// match reports whether all constrained parameters match.
func (c conditions) match(params httprouter.Params) bool {
	for k, r := range c {
		v := params[k].Value
		if !r.MatchString(v) {
			return false
		}
	}
	return true
}
