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

type Route struct {
	router          *Router
	methods         []string
	pattern         pattern
	paramNames      helpers.Slice[string]
	paramNamesMatch [][]string
	conditions      conditions
	handler         http.Handler
}

// Name устанавливает имя маршрута.
func (route *Route) Name(name string) *Route {
	route.router.routeByName[name] = route
	return route
}

// Where задаёт регулярное выражение для проверки именованного параметра.
func (route *Route) Where(param string, regexp *regexp.Regexp) *Route {
	i := route.paramNames.IndexOf(param)
	if i < 0 {
		panic("unknown parameter: " + param)
	}

	route.conditions[i] = regexp
	return route
}

// Handle устанавливает обработчик маршрута.
func (route *Route) Handle(handler http.Handler) *Route {
	route.handler = handler
	return route
}

// Handle устанавливает обработчик маршрута.
func (route *Route) HandleFunc(handlerFunc http.HandlerFunc) *Route {
	return route.Handle(handlerFunc)
}

// Url генерирует URL адрес для данного маршрута.
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
