package router

import (
	"net/http"
)

// defaultRouter is the router used by the package-level functions.
var (
	defaultRouter *Router
)

// DefaultRouter returns the shared router used by the package-level
// functions, creating it on first use.
func DefaultRouter() *Router {
	if defaultRouter == nil {
		defaultRouter = New()
	}
	return defaultRouter
}

// ParseMap adds routes from a nested map to the default router.
// See Router.ParseMap.
func ParseMap(
	m map[string]interface{},
	handlerByName func(string) http.Handler,
	middlewareByName func(string) MiddlewareFunc,
) {
	DefaultRouter().ParseMap(m, handlerByName, middlewareByName)
}

// Group adds a route group to the default router. See Router.Group.
func Group(f func(*Router)) {
	DefaultRouter().Group(f)
}

// Prefix adds routes under a path prefix to the default router.
// See Router.Prefix.
func Prefix(path string, f func(*Router)) {
	DefaultRouter().Prefix(path, f)
}

// Use adds middleware to the default router. See Router.Use.
func Use(middleware ...MiddlewareFunc) {
	DefaultRouter().Use(middleware...)
}

// Get adds a GET route to the default router.
func Get(path string) *Route {
	return DefaultRouter().Get(path)
}

// Post adds a POST route to the default router.
func Post(path string) *Route {
	return DefaultRouter().Post(path)
}

// Put adds a PUT route to the default router.
func Put(path string) *Route {
	return DefaultRouter().Put(path)
}

// Patch adds a PATCH route to the default router.
func Patch(path string) *Route {
	return DefaultRouter().Patch(path)
}

// Delete adds a DELETE route to the default router.
func Delete(path string) *Route {
	return DefaultRouter().Delete(path)
}

// Options adds an OPTIONS route to the default router.
func Options(path string) *Route {
	return DefaultRouter().Options(path)
}

// NewRoute adds a route for the given methods to the default router.
func NewRoute(path string, methods ...string) *Route {
	return DefaultRouter().NewRoute(path, methods...)
}

// Url builds the URL path of a named route of the default router.
// See Route.Url.
func Url(name string, params ...interface{}) (string, error) {
	return DefaultRouter().Url(name, params...)
}

// HandleNotFound sets the not-found handler of the default router.
func HandleNotFound(handler http.Handler) {
	DefaultRouter().HandleNotFound(handler)
}

// HandleMethodNotAllowed sets the method-not-allowed handler of the default
// router.
func HandleMethodNotAllowed(handler http.Handler) {
	DefaultRouter().HandleMethodNotAllowed(handler)
}

// HandlePanic sets the panic handler of the default router.
// See Router.HandlePanic.
func HandlePanic(handler func(http.ResponseWriter, *http.Request, interface{})) {
	DefaultRouter().HandlePanic(handler)
}
