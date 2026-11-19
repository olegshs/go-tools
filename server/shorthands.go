package server

import (
	"net/http"

	"github.com/olegshs/go-tools/config"
)

// Start creates a server from the "server" config section and runs it.
// It blocks until the server fails.
func Start(handler http.Handler, middleware ...MiddlewareFunc) error {
	conf := DefaultConfig()
	config.GetStruct("server", &conf)

	s := NewServer(conf, handler, middleware...)
	return s.Start()
}
