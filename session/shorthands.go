package session

import (
	"net/http"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/events"
)

var (
	defaultManager ManagerInterface
)

func DefaultManager() ManagerInterface {
	if defaultManager == nil {
		defaultManager = newDefaultManager()
	}
	return defaultManager
}

func newDefaultManager() ManagerInterface {
	conf := DefaultConfig()
	config.GetStruct("session", &conf)

	p := NewManager(conf, newStorage(conf))
	return p
}

func Start(r *http.Request, w http.ResponseWriter) SessionInterface {
	return DefaultManager().Start(r, w)
}

func Restart(s SessionInterface, w http.ResponseWriter) {
	DefaultManager().Restart(s, w)
}

func Open(id string) SessionInterface {
	return DefaultManager().Open(id)
}

func Id(r *http.Request) string {
	return DefaultManager().Id(r)
}

func SetId(w http.ResponseWriter, id string) {
	DefaultManager().SetId(w, id)
}

func GenerateId() string {
	return DefaultManager().GenerateId()
}

func Events() *events.Dispatcher {
	return DefaultManager().Events()
}
