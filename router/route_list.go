package router

import (
	"github.com/julienschmidt/httprouter"
)

// routeList holds the routes registered for one method and httprouter path,
// in the order they were added.
type routeList []*Route

// match returns the first route that has a handler and whose constraints
// match params, or nil.
func (routes *routeList) match(params httprouter.Params) *Route {
	for _, route := range *routes {
		if route.handler == nil {
			continue
		}
		if route.conditions.match(params) {
			return route
		}
	}
	return nil
}
