package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/helpers/typeconv"
)

// Route is a single route. Its methods return the route so calls can be
// chained.
type Route struct {
	router     *Router
	methods    []string
	pattern    pattern
	paramNames helpers.Slice[string]
	// paramNamesMatch holds the regexp submatches for each parameter
	// placeholder, used to substitute values in Url.
	paramNamesMatch [][]string
	// conditions maps parameter positions to their constraints.
	conditions conditions
	handler    http.Handler
}

// Name sets the route name used by Router.Url. A later route with the same
// name replaces the earlier one.
func (route *Route) Name(name string) *Route {
	route.router.routeByName[name] = route
	return route
}

// Where constrains the parameter to values matching regexp. A request whose
// value does not match falls through to the next route with the same path
// shape, or to the not-found handler. It panics if the route has no such
// parameter.
func (route *Route) Where(param string, regexp *regexp.Regexp) *Route {
	i := route.paramNames.IndexOf(param)
	if i < 0 {
		panic("unknown parameter: " + param)
	}

	route.conditions[i] = regexp
	return route
}

// Handle sets the route handler.
func (route *Route) Handle(handler http.Handler) *Route {
	route.handler = handler
	return route
}

// HandleFunc sets the route handler to a function.
func (route *Route) HandleFunc(handlerFunc http.HandlerFunc) *Route {
	return route.Handle(handlerFunc)
}

// Url builds the URL path of the route by replacing the parameters with
// params, in order. Values are converted to strings and checked against the
// constraints. Extra values are appended as additional path segments.
// It returns ErrNotEnoughParameters or ErrInvalidParameter on failure.
// Values are not URL-escaped.
func (route *Route) Url(params ...interface{}) (string, error) {
	nParams := len(params)
	nMatch := len(route.paramNamesMatch)
	if nParams < nMatch {
		err := fmt.Errorf("%w (%d < %d)",
			ErrNotEnoughParameters, nParams, nMatch,
		)
		return "", err
	}

	u := string(route.pattern)

	for i, v := range params {
		s := typeconv.String(v)

		if (route.conditions[i] != nil) && !route.conditions[i].MatchString(s) {
			err := fmt.Errorf("%w: %s not match %s",
				ErrInvalidParameter, strconv.Quote(s), route.conditions[i],
			)
			return "", err
		}

		if i < nMatch {
			m := route.paramNamesMatch[i]
			u = strings.ReplaceAll(u, m[0], s)
		} else {
			u += "/" + s
		}
	}

	return u, nil
}

// namedParams converts httprouter's positional parameters to Params with
// the names from the route pattern.
func (route *Route) namedParams(params httprouter.Params) Params {
	n := len(params)
	if n == 0 {
		return nil
	}

	named := make(Params, n)
	for i, param := range params {
		named[i] = Param{
			Key:   route.paramNames[i],
			Value: param.Value,
		}
	}

	return named
}
