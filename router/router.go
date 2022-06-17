// Пакет router реализует маршрутизатор HTTP запросов.
//
// Является обёрткой над пакетом "github.com/julienschmidt/httprouter".
// Добавляет различные возможности, такие как: генерация URL для именованных маршрутов,
// префиксы и группы маршрутов, функции-посредники (middleware),
// проверка именованных параметров с помощью регулярных выражений.
package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type Router struct {
	prefix      pattern
	conditions  conditions
	middleware  middlewareList
	routes      routeMap
	routeByName map[string]*Route
	r           *httprouter.Router
}

// New создаёт новый экземпляр маршрутизатора.
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

// ParseMap добавляет маршруты, заданные в карте со специальной структурой (см. пример).
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

// Group добавляет группу маршрутов.
// Для группы могут быть заданы функции-посредники.
func (router *Router) Group(f func(*Router)) {
	f(router.clone())
}

// Prefix добавляет группу маршрутов с заданным префиксом.
// Префикс может содержать именованные параметры.
func (router *Router) Prefix(path string, f func(*Router)) {
	sub := router.clone()
	sub.prefix = router.prefix + pattern(path)

	f(sub)
}

// Use добавляет функции-посредники,
// которые будут использоваться маршрутизатором или группой маршрутов.
func (router *Router) Use(middleware ...MiddlewareFunc) {
	router.middleware = append(router.middleware, middleware...)
}

// Where задаёт регулярное выражение для проверки именованного параметра,
// заданного в префиксе.
func (router *Router) Where(param string, regexp *regexp.Regexp) {
	i := router.prefix.paramNames().IndexOf(param)
	if i < 0 {
		panic("unknown parameter: " + param)
	}

	router.conditions[i] = regexp
}

// Get создаёт и возвращает маршрут для обработки запросов GET.
func (router *Router) Get(path string) *Route {
	return router.NewRoute(path, http.MethodGet)
}

// Post создаёт и возвращает маршрут для обработки запросов POST.
func (router *Router) Post(path string) *Route {
	return router.NewRoute(path, http.MethodPost)
}

// Put создаёт и возвращает маршрут для обработки запросов PUT.
func (router *Router) Put(path string) *Route {
	return router.NewRoute(path, http.MethodPut)
}

// Patch создаёт и возвращает маршрут для обработки запросов PATCH.
func (router *Router) Patch(path string) *Route {
	return router.NewRoute(path, http.MethodPatch)
}

// Delete создаёт и возвращает маршрут для обработки запросов DELETE.
func (router *Router) Delete(path string) *Route {
	return router.NewRoute(path, http.MethodDelete)
}

// Options создаёт и возвращает маршрут для обработки запросов OPTIONS.
func (router *Router) Options(path string) *Route {
	return router.NewRoute(path, http.MethodOptions)
}

// NewRoute создаёт и возвращает маршрут для обработки запросов, отправленных указанными методами.
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

// Url генерирует URL адрес для именованного маршрута.
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

// HandleNotFound устанавливает обработчик, который вызывается если маршрут не найден.
func (router *Router) HandleNotFound(handler http.Handler) {
	router.r.NotFound = router.middleware.wrap(handler)
}

// HandleMethodNotAllowed устанавливает обработчик, который вызывается если маршрут найден,
// но метод запроса не поддерживается.
func (router *Router) HandleMethodNotAllowed(handler http.Handler) {
	router.r.MethodNotAllowed = router.middleware.wrap(handler)
}

// HandlePanic устанавливает обработчик panic при вызове маршрута.
// Обработчику передаются http.ResponseWriter, *http.Request,
// а также значение возвращаемое функцией recover.
func (router *Router) HandlePanic(handler func(http.ResponseWriter, *http.Request, interface{})) {
	router.r.PanicHandler = handler
}

// ServeHTTP реализует интерфейс http.Handler.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.r.ServeHTTP(w, r)
}

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

func (router *Router) newHandler(routes *routeList) http.Handler {
	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := httprouter.ParamsFromContext(r.Context())
		for i, param := range params {
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
