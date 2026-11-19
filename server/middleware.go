package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/logs"
)

const (
	// The header added to every response when Config.Tokens is set.
	headerName  = "X-Powered-By"
	headerValue = "go-tools"
)

// MiddlewareFunc wraps a handler with extra behavior and returns the new
// handler.
type MiddlewareFunc func(http.Handler) http.Handler

// addMiddleware builds the handler chain described in the package
// documentation.
func addMiddleware(conf Config, handler http.Handler, middleware ...MiddlewareFunc) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		m := middleware[i]
		handler = m(handler)
	}

	if conf.Static.Enabled {
		if conf.Log.Enabled {
			logs.Channel(conf.Log.Channel).Notice(fmt.Sprintf("Static assets directory is %s\n", conf.Static.Dir))
		}

		handler = staticMiddleware(handler, conf.Static)
	}

	if conf.AccessLog.Enabled {
		handler = loggingMiddleware(handler, conf.AccessLog)
	}

	if conf.Tokens {
		handler = tokensMiddleware(handler)
	}

	return handler
}

func tokensMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerName, headerValue)
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler, conf ConfigAccessLog) http.Handler {
	log := logs.Channel(conf.Channel)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info(r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

// staticMiddleware serves existing files from the static directory and
// passes all other requests to next.
func staticMiddleware(next http.Handler, conf ConfigStatic) http.Handler {
	dir := config.AbsPath(conf.Dir)

	fileServer := http.NewServeMux()
	fileServer.Handle("/", http.FileServer(http.Dir(dir)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isValidStaticRequest(r, dir) {
			fileServer.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// isValidStaticRequest reports whether the request path names a file in dir,
// or a directory that contains index.html. Paths containing "/." are
// refused, which blocks hidden files and ".." segments.
func isValidStaticRequest(r *http.Request, dir string) bool {
	if strings.Contains(r.URL.Path, "/.") {
		return false
	}

	fi, err := os.Stat(dir + r.URL.Path)
	if err != nil {
		return false
	}

	if fi.IsDir() {
		fi, err := os.Stat(dir + r.URL.Path + "/index.html")
		if err != nil || fi.IsDir() {
			return false
		}
	}

	return true
}
