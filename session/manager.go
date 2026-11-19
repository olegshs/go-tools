package session

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/olegshs/go-tools/events"
	"github.com/olegshs/go-tools/session/storage"
)

// Manager creates and opens sessions and manages the session cookie.
type Manager struct {
	conf    Config
	storage StorageInterface
	events  *events.Dispatcher
}

// NewManager creates a manager that keeps session data in stor.
// The storage's EventBeforeDelete is re-dispatched on the manager as
// EventBeforeDestroy.
func NewManager(conf Config, stor StorageInterface) ManagerInterface {
	mgr := new(Manager)
	mgr.conf = conf
	mgr.storage = stor
	mgr.events = events.New()

	mgr.storage.Events().AddListener(storage.EventBeforeDelete, func(id string) {
		mgr.events.Dispatch(EventBeforeDestroy, id)
	})

	return mgr
}

// Start opens the session of the request. If the request has no valid
// session cookie, it generates a new ID and sets the cookie on w, so Start
// must be called before the response headers are written.
func (mgr *Manager) Start(r *http.Request, w http.ResponseWriter) SessionInterface {
	id := mgr.Id(r)

	if id == "" {
		id = mgr.GenerateId()
		mgr.SetId(w, id)
	}

	s := mgr.Open(id)
	return s
}

// Restart clears the session, deletes it from storage, and gives it a new ID
// that is sent to the client in a new cookie. Use it on login and logout to
// prevent session fixation.
func (mgr *Manager) Restart(s SessionInterface, w http.ResponseWriter) {
	s.DeleteAll()
	s.Destroy()

	id := mgr.GenerateId()
	s.SetId(id)
	mgr.SetId(w, id)
}

// Open loads the session with the given ID. It is not checked that the
// session exists; a missing session is opened empty.
func (mgr *Manager) Open(id string) SessionInterface {
	s := mgr.newSession(id)
	s.load()

	return s
}

// Id returns the session ID from the request cookie, or "" if there is no
// cookie or its value has the wrong length.
func (mgr *Manager) Id(r *http.Request) string {
	c, err := r.Cookie(mgr.conf.Cookie.Name)
	if err != nil {
		return ""
	}

	id := c.Value

	length := mgr.conf.Id.Length
	if len(id) != hex.EncodedLen(length) {
		return ""
	}

	return id
}

// SetId sets the session cookie on w. The cookie is HttpOnly and has no
// expiry time, so it lasts until the browser is closed.
func (mgr *Manager) SetId(w http.ResponseWriter, id string) {
	c := new(http.Cookie)
	c.Name = mgr.conf.Cookie.Name
	c.Value = id
	c.Path = mgr.conf.Cookie.Path
	c.Domain = mgr.conf.Cookie.Domain
	c.Secure = mgr.conf.Cookie.Secure
	c.HttpOnly = true

	http.SetCookie(w, c)
}

// GenerateId returns a new random session ID: Config.Id.Length random bytes
// in hex. It panics if the system random source fails.
func (mgr *Manager) GenerateId() string {
	length := mgr.conf.Id.Length

	r := make([]byte, length)

	_, err := rand.Read(r)
	if err != nil {
		panic(err)
	}

	id := hex.EncodeToString(r)
	return id
}

// Events returns the dispatcher for session events (EventAfterLoad,
// EventBeforeSave and EventBeforeDestroy).
func (mgr *Manager) Events() *events.Dispatcher {
	return mgr.events
}

func (mgr *Manager) newSession(id string) *Session {
	s := new(Session)
	s.storage = mgr.storage
	s.events = mgr.events
	s.id = id

	return s
}
