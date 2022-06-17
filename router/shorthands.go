package router

import (
	"net/http"
)

var (
	defaultRouter *Router
)

// DefaultRouter возвращает используемый по умолчанию экземпляр маршрутизатора.
func DefaultRouter() *Router {
	if defaultRouter == nil {
		defaultRouter = New()
	}
	return defaultRouter
}

// ParseMap добавляет маршруты, заданные в карте со специальной структурой (см. пример).
func ParseMap(
	m map[string]interface{},
	handlerByName func(string) http.Handler,
	middlewareByName func(string) MiddlewareFunc,
) {
	DefaultRouter().ParseMap(m, handlerByName, middlewareByName)
}

// Group добавляет группу маршрутов.
// Для группы могут быть заданы функции-посредники.
func Group(f func(*Router)) {
	DefaultRouter().Group(f)
}

// Prefix добавляет группу маршрутов с заданным префиксом.
// Префикс может содержать именованные параметры.
func Prefix(path string, f func(*Router)) {
	DefaultRouter().Prefix(path, f)
}

// Use добавляет функции-посредники,
// которые будут использоваться маршрутизатором или группой маршрутов.
func Use(middleware ...MiddlewareFunc) {
	DefaultRouter().Use(middleware...)
}

// Get создаёт и возвращает маршрут для обработки запросов GET.
func Get(path string) *Route {
	return DefaultRouter().Get(path)
}

// Post создаёт и возвращает маршрут для обработки запросов POST.
func Post(path string) *Route {
	return DefaultRouter().Post(path)
}

// Put создаёт и возвращает маршрут для обработки запросов PUT.
func Put(path string) *Route {
	return DefaultRouter().Put(path)
}

// Patch создаёт и возвращает маршрут для обработки запросов PATCH.
func Patch(path string) *Route {
	return DefaultRouter().Patch(path)
}

// Delete создаёт и возвращает маршрут для обработки запросов DELETE.
func Delete(path string) *Route {
	return DefaultRouter().Delete(path)
}

// Options создаёт и возвращает маршрут для обработки запросов OPTIONS.
func Options(path string) *Route {
	return DefaultRouter().Options(path)
}

// NewRoute создаёт и возвращает маршрут для обработки запросов, отправленных указанными методами.
func NewRoute(path string, methods ...string) *Route {
	return DefaultRouter().NewRoute(path, methods...)
}

// Url генерирует URL адрес для именованного маршрута.
func Url(name string, params ...interface{}) (string, error) {
	return DefaultRouter().Url(name, params...)
}

// HandleNotFound устанавливает обработчик, который вызывается если маршрут не найден.
func HandleNotFound(handler http.Handler) {
	DefaultRouter().HandleNotFound(handler)
}

// HandleMethodNotAllowed устанавливает обработчик, который вызывается если маршрут найден,
// но метод запроса не поддерживается.
func HandleMethodNotAllowed(handler http.Handler) {
	DefaultRouter().HandleMethodNotAllowed(handler)
}

// HandlePanic устанавливает обработчик panic при вызове маршрута.
// Обработчику передаются http.ResponseWriter, *http.Request,
// а также значение возвращаемое функцией recover.
func HandlePanic(handler func(http.ResponseWriter, *http.Request, interface{})) {
	DefaultRouter().HandlePanic(handler)
}
