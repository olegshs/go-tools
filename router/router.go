// Package router implements an HTTP request router on top of
// github.com/julienschmidt/httprouter.
//
// It adds named routes with URL generation, route groups and path prefixes,
// middleware per group, and regular expression constraints on path
// parameters.
//
// Path parameters are written as {name}. A catch-all parameter that matches
// the rest of the path is written as {name...}. Parameter values reach
// handlers through ParamsFromRequest:
//
//	r := router.New()
//	r.Get("/articles/{id}").
//		Where("id", regexp.MustCompile(`^\d+$`)).
//		Name("articles.get").
//		HandleFunc(func(w http.ResponseWriter, r *http.Request) {
//			id := router.ParamsFromRequest(r).ByName("id")
//			...
//		})
//	u, _ := r.Url("articles.get", 42) // "/articles/42"
//
// Several routes may share the same method and path shape and differ only
// in their constraints. They are tried in the order they were added, and the
// first one whose constraints all match handles the request.
//
// Routes can also be built from a nested map, such as one loaded from a
// config file. See Router.ParseMap.
package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Router registers routes and dispatches requests to them. It implements
// http.Handler.
//
// The *Router passed to Group and Prefix callbacks shares the route table
// with its parent but has its own prefix, constraints and middleware.
// Routes must be registered before the router starts serving requests.
type Router struct {
	prefix      pattern
	conditions  conditions
	middleware  middlewareList
	routes      routeMap
	routeByName map[string]*Route
	r           *httprouter.Router
}

// New creates an empty router. Unmatched requests get http.NotFoundHandler.
func New() *Router {
	router := new(Router)
	router.prefix = ""
	router.conditions = make(conditions)
	router.middleware = make(middlewareList, 0)
	router.routes = make(routeMap)
	router.routeByName = make(map[string]*Route)
	router.r = httprouter.New()
	router.r.NotFound = http.NotFoundHandler()

	return router
}

// ParseMap adds routes described by a nested map. Keys are handled in sorted
// order, so "$" directives come before the routes they apply to.
//
//   - "GET /path", "GET, POST /path" or just "GET": a route relative to the
//     current prefix. The value is either the route name or a map with a
//     "$name" key and parameter name to regexp pairs.
//   - "/prefix": the value is a map of routes under that prefix. The prefix
//     may contain parameters.
//   - "(any text)": the value is a map of routes in a group without a prefix,
//     typically used to apply middleware to some routes.
//   - "$use": a middleware name or a list of names, applied to the
//     current group.
//   - "$where": a map of parameter name to regexp for parameters in the
//     current prefix.
//
// The route name is used both for URL generation and to look up the handler
// through handlerByName. Middleware names are resolved through
// middlewareByName, and names it returns nil for are ignored. Invalid
// regexps cause a panic. See ExampleRouter_ParseMap.
func (router *Router) ParseMap(
	m map[string]interface{},
	handlerByName func(string) http.Handler,
	middlewareByName func(string) MiddlewareFunc,
) {
	p := &parser{
		router:           router,
		handlerByName:    handlerByName,
		middlewareByName: middlewareByName,
	}
	p.ParseMap(m)
}

// Group calls f with a sub-router that shares this router's prefix,
// constraints and middleware. Middleware added in f with Use applies only to
// routes registered inside f.
func (router *Router) Group(f func(*Router)) {
	f(router.clone())
}

// Prefix calls f with a sub-router whose routes start with path. The prefix
// may contain parameters, which can be constrained with the sub-router's
// Where.
func (router *Router) Prefix(path string, f func(*Router)) {
	sub := router.clone()
	sub.prefix = router.prefix + pattern(path)

	f(sub)
}

// Use adds middleware to the router or group. It applies to routes added
// afterwards. The first middleware added is the outermost.
func (router *Router) Use(middleware ...MiddlewareFunc) {
	router.middleware = append(router.middleware, middleware...)
}

// Where constrains a parameter of the current prefix to values matching
// regexp. It applies to routes added afterwards and panics if the prefix has
// no such parameter.
func (router *Router) Where(param string, regexp *regexp.Regexp) {
	i := router.prefix.paramNames().IndexOf(param)
	if i < 0 {
		panic("unknown parameter: " + param)
	}

	router.conditions[i] = regexp
}

// Get adds a route for GET requests.
func (router *Router) Get(path string) *Route {
	return router.NewRoute(path, http.MethodGet)
}

// Post adds a route for POST requests.
func (router *Router) Post(path string) *Route {
	return router.NewRoute(path, http.MethodPost)
}

// Put adds a route for PUT requests.
func (router *Router) Put(path string) *Route {
	return router.NewRoute(path, http.MethodPut)
}

// Patch adds a route for PATCH requests.
func (router *Router) Patch(path string) *Route {
	return router.NewRoute(path, http.MethodPatch)
}

// Delete adds a route for DELETE requests.
func (router *Router) Delete(path string) *Route {
	return router.NewRoute(path, http.MethodDelete)
}

// Options adds a route for OPTIONS requests.
func (router *Router) Options(path string) *Route {
	return router.NewRoute(path, http.MethodOptions)
}

// NewRoute adds a route for the given HTTP methods. The path is appended to
// the current prefix and inherits the prefix constraints. The route serves
// nothing until a handler is set with Handle or HandleFunc.
func (router *Router) NewRoute(path string, methods ...string) *Route {
	route := new(Route)
	route.router = router
	route.methods = methods
	route.pattern = router.prefix + pattern(path)
	route.paramNames = route.pattern.paramNames()
	route.paramNamesMatch = route.pattern.paramNamesMatch()
	route.conditions = router.conditions.clone()

	router.addRoute(route)

	return route
}

// Url builds the URL path of the named route. See Route.Url.
// It returns ErrRouteNotFound if there is no route with that name.
func (router *Router) Url(name string, params ...interface{}) (string, error) {
	route, ok := router.routeByName[name]
	if !ok {
		return "", fmt.Errorf("%s: %w", name, ErrRouteNotFound)
	}

	u, err := route.Url(params...)
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}

	return u, nil
}

// HandleNotFound sets the handler for requests that match no route. It is
// wrapped in the middleware added so far with Use.
func (router *Router) HandleNotFound(handler http.Handler) {
	router.r.NotFound = router.middleware.wrap(handler)
}

// HandleMethodNotAllowed sets the handler for requests whose path matches a
// route that does not accept the request method. It is wrapped in the
// middleware added so far with Use.
func (router *Router) HandleMethodNotAllowed(handler http.Handler) {
	router.r.MethodNotAllowed = router.middleware.wrap(handler)
}

// HandlePanic sets a handler that is called when a route handler panics.
// It receives the response writer, the request and the value returned by
// recover.
func (router *Router) HandlePanic(handler func(http.ResponseWriter, *http.Request, interface{})) {
	router.r.PanicHandler = handler
}

// ServeHTTP implements http.Handler.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.r.ServeHTTP(w, r)
}

// clone returns a sub-router that shares the route table and the underlying
// httprouter but has its own copies of the prefix, constraints and
// middleware.
func (router *Router) clone() *Router {
	clone := new(Router)
	clone.prefix = router.prefix
	clone.conditions = router.conditions.clone()
	clone.middleware = router.middleware.clone()
	clone.routes = router.routes
	clone.routeByName = router.routeByName
	clone.r = router.r

	return clone
}

// addRoute registers route for each of its methods. httprouter allows only
// one handler per method and path, so routes with the same path shape are
// collected in a routeList that is registered once. That handler uses the
// middleware of the router that added the first route for that path.
func (router *Router) addRoute(route *Route) {
	p := route.pattern.httpRouterString()

	for _, method := range route.methods {
		a := router.routes.get(method, p)

		if len(*a) == 0 {
			h := router.newHandler(a)
			router.r.Handler(method, p, h)
		}

		*a = append(*a, route)
	}
}

// newHandler returns the httprouter handler for a list of routes with the
// same path shape. It picks the first route whose constraints match, stores
// the named parameters in the request context and calls the route handler.
func (router *Router) newHandler(routes *routeList) http.Handler {
	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := httprouter.ParamsFromContext(r.Context())
		for i, param := range params {
			// Catch-all values start with "/"; strip it so all values look alike.
			params[i].Value = strings.Trim(param.Value, "/")
		}

		route := routes.match(params)
		if route == nil {
			router.r.NotFound.ServeHTTP(w, r)
			return
		}

		namedParams := route.namedParams(params)
		if len(namedParams) > 0 {
			namedParams.toRequest(r)
		}

		route.handler.ServeHTTP(w, r)
	})

	handler = router.middleware.wrap(handler)

	return handler
}
