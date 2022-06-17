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
	headerName  = "X-Powered-By"
	headerValue = "goweb"
)

type MiddlewareFunc func(http.Handler) http.Handler

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

func staticMiddleware(next http.Handler, conf ConfigStatic) http.Handler {
	fileServer := http.NewServeMux()
	fileServer.Handle("/", http.FileServer(http.Dir(config.AbsPath(conf.Dir))))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isValidStaticRequest(r, conf.Dir) {
			fileServer.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

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
