package session

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/olegshs/go-tools/events"
	"github.com/olegshs/go-tools/session/storage"
)

type Manager struct {
	conf    Config
	storage StorageInterface
	events  *events.Dispatcher
}

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

func (mgr *Manager) Start(r *http.Request, w http.ResponseWriter) SessionInterface {
	id := mgr.Id(r)

	if id == "" {
		id = mgr.GenerateId()
		mgr.SetId(w, id)
	}

	s := mgr.Open(id)
	return s
}

func (mgr *Manager) Restart(s SessionInterface, w http.ResponseWriter) {
	s.DeleteAll()
	s.Destroy()

	id := mgr.GenerateId()
	s.SetId(id)
	mgr.SetId(w, id)
}

func (mgr *Manager) Open(id string) SessionInterface {
	s := mgr.newSession(id)
	s.load()

	return s
}

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
