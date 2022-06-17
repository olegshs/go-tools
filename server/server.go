// Пакет server предоставляет обёртку для работы с http.Server.
package server

import (
	"fmt"
	"net/http"

	"github.com/olegshs/go-tools/logs"
)

type Server struct {
	conf       Config
	handler    http.Handler
	middleware []MiddlewareFunc
}

func NewServer(conf Config, handler http.Handler, middleware ...MiddlewareFunc) *Server {
	s := new(Server)
	s.conf = conf
	s.handler = handler
	s.middleware = middleware

	return s
}

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
