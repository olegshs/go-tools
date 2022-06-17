package server

import (
	"net/http"

	"github.com/olegshs/go-tools/config"
)

func Start(handler http.Handler, middleware ...MiddlewareFunc) error {
	conf := DefaultConfig()
	config.GetStruct("server", &conf)

	s := NewServer(conf, handler, middleware...)
	return s.Start()
}
