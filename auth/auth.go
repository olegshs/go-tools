// Пакет auth предоставляет средства для аутентификации и авторизации.
package auth

import (
	"net/http"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/session"
)

var (
	UserFactory UserFactoryInterface = new(dummyUserFactory)

	defaultManager ManagerInterface
)

type TokenPayload struct {
	UserId     int64
	Time       uint32
	RememberIP bool
}

func DefaultManager() ManagerInterface {
	if defaultManager == nil {
		defaultManager = newDefaultManager()
	}
	return defaultManager
}

func newDefaultManager() ManagerInterface {
	conf := DefaultConfig()
	config.GetStruct("auth", &conf)

	p := NewManager(conf, UserFactory, session.DefaultManager())
	return p
}

func UserId(s session.SessionInterface) int64 {
	return DefaultManager().UserId(s)
}

func SetUserId(s session.SessionInterface, id int64) {
	DefaultManager().SetUserId(s, id)
}

func User(s session.SessionInterface) UserInterface {
	return DefaultManager().User(s)
}

func Authenticate(login string, password string) bool {
	return DefaultManager().Authenticate(login, password)
}

func Login(s session.SessionInterface, w http.ResponseWriter, login string, password string) bool {
	return DefaultManager().Login(s, w, login, password)
}

func Logout(s session.SessionInterface, w http.ResponseWriter) {
	DefaultManager().Logout(s, w)
}

func Remember(s session.SessionInterface, r *http.Request, w http.ResponseWriter, rememberIP bool) {
	DefaultManager().Remember(s, r, w, rememberIP)
}

func Restore(s session.SessionInterface, r *http.Request, w http.ResponseWriter) UserInterface {
	return DefaultManager().Restore(s, r, w)
}

func PasswordHash(password string, salt string) string {
	return DefaultManager().PasswordHash(password, salt)
}
