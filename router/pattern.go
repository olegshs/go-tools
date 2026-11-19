package router

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/olegshs/go-tools/helpers"
)

// pattern is a route path with {name} and {name...} parameter placeholders.
type pattern string

var (
	// paramRegexp matches a parameter placeholder. Group 1 is the name and
	// group 2 is "..." for catch-all parameters.
	paramRegexp = regexp.MustCompile(`{([A-Za-z_][0-9A-Za-z_]*)(\.\.\.)?}`)
)

// paramNames returns the parameter names in order.
func (p pattern) paramNames() helpers.Slice[string] {
	a := p.paramNamesMatch()
	names := make([]string, len(a))

	for i, m := range a {
		names[i] = m[1]
	}

	return names
}

// paramNamesMatch returns the submatches of paramRegexp for each parameter.
func (p pattern) paramNamesMatch() [][]string {
	return paramRegexp.FindAllStringSubmatch(string(p), -1)
}

// httpRouterString converts the pattern to httprouter syntax. Parameters
// become :0, :1, ... and catch-alls become *N, so routes that differ only
// in parameter names map to the same httprouter path. Literal ':' and '*'
// are removed because httprouter would treat them as parameters.
func (p pattern) httpRouterString() string {
	s := string(p)
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "*", "")

	a := paramRegexp.FindAllStringSubmatch(s, -1)

	for i, m := range a {
		var repl string
		if m[2] == "..." {
			repl = fmt.Sprintf("*%d", i)
		} else {
			repl = fmt.Sprintf(":%d", i)
		}

		s = strings.ReplaceAll(s, m[0], repl)
	}

	return s
}
