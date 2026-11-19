package session

import (
	"net/http"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/events"
)

// defaultManager is the manager used by the package-level functions.
var (
	defaultManager ManagerInterface
)

// DefaultManager returns the shared session manager, creating it from the
// "session" config section on first use. It panics if the configured
// storage driver is unknown.
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

// Start opens the session of the request with the default manager.
// See Manager.Start.
func Start(r *http.Request, w http.ResponseWriter) SessionInterface {
	return DefaultManager().Start(r, w)
}

// Restart gives the session a new ID with the default manager.
// See Manager.Restart.
func Restart(s SessionInterface, w http.ResponseWriter) {
	DefaultManager().Restart(s, w)
}

// Open loads a session by ID with the default manager.
func Open(id string) SessionInterface {
	return DefaultManager().Open(id)
}

// Id returns the session ID from the request cookie.
func Id(r *http.Request) string {
	return DefaultManager().Id(r)
}

// SetId sets the session cookie on w.
func SetId(w http.ResponseWriter, id string) {
	DefaultManager().SetId(w, id)
}

// GenerateId returns a new random session ID.
func GenerateId() string {
	return DefaultManager().GenerateId()
}

// Events returns the dispatcher of the default manager.
func Events() *events.Dispatcher {
	return DefaultManager().Events()
}
