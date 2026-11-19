// Package server runs an http.Server configured from the "server" config
// section, with optional static file serving and access logging.
//
//	server.Start(router.DefaultRouter())
//
// The handler chain is, from outermost to innermost: the X-Powered-By
// header, the access log, static files, the middleware passed to Start,
// and the handler. Static files are served without running that middleware.
package server

import (
	"fmt"
	"net/http"

	"github.com/olegshs/go-tools/logs"
)

// Server is an HTTP server with a handler and middleware.
type Server struct {
	conf       Config
	handler    http.Handler
	middleware []MiddlewareFunc
}

// NewServer creates a server. The first middleware is the outermost.
func NewServer(conf Config, handler http.Handler, middleware ...MiddlewareFunc) *Server {
	s := new(Server)
	s.conf = conf
	s.handler = handler
	s.middleware = middleware

	return s
}

// Start listens on the configured address and serves requests. It blocks
// and returns the error from http.Server.ListenAndServe.
func (s *Server) Start() error {
	if s.conf.Log.Enabled {
		logs.Channel(s.conf.Log.Channel).Notice(fmt.Sprintf("Listening %s\n", s.conf.Listen))
	}

	server := &http.Server{
		Addr:              s.conf.Listen,
		ReadTimeout:       s.conf.ReadTimeout,
		ReadHeaderTimeout: s.conf.ReadHeaderTimeout,
		WriteTimeout:      s.conf.WriteTimeout,
		IdleTimeout:       s.conf.IdleTimeout,
		MaxHeaderBytes:    s.conf.MaxHeaderBytes,
	}
	server.Handler = addMiddleware(s.conf, s.handler, s.middleware...)

	return server.ListenAndServe()
}
