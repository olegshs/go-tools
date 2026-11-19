// Package auth implements password login and "remember me" cookies on top
// of package session.
//
// The application provides its users by setting UserFactory at startup.
// The logged-in user's ID is stored in the session under
// Config.Session.UserIdKey:
//
//	auth.UserFactory = myUsers{}
//
//	s := session.Start(r, w)
//	defer s.Close()
//	if auth.Login(s, w, login, password) {
//		auth.Remember(s, r, w, false)
//	}
//	user := auth.User(s)
//	if user == nil {
//		user = auth.Restore(s, r, w) // log in from the remember-me cookie
//	}
//
// Settings are read from the "auth" config section. Set cookie.secret to a
// long random value, because it signs the remember-me cookies.
package auth

import (
	"net/http"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/session"
)

var (
	// UserFactory loads users for the default manager. It must be set before
	// the first call to a package-level function. The default factory finds no
	// users.
	UserFactory UserFactoryInterface = new(dummyUserFactory)

	defaultManager ManagerInterface
)

// TokenPayload is the content of a remember-me cookie. Time is the Unix time
// when the cookie was issued. RememberIP binds the cookie to the client IP
// address.
type TokenPayload struct {
	UserId     int64
	Time       uint32
	RememberIP bool
}

// DefaultManager returns the shared auth manager, creating it on first use
// from the "auth" config section, UserFactory and the default session
// manager.
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

// UserId returns the ID of the user logged in to the session, or 0.
func UserId(s session.SessionInterface) int64 {
	return DefaultManager().UserId(s)
}

// SetUserId marks the session as logged in as the user with the given ID.
func SetUserId(s session.SessionInterface, id int64) {
	DefaultManager().SetUserId(s, id)
}

// User returns the user logged in to the session, or nil.
func User(s session.SessionInterface) UserInterface {
	return DefaultManager().User(s)
}

// Authenticate reports whether the login and password are valid.
// See Manager.Authenticate.
func Authenticate(login string, password string) bool {
	return DefaultManager().Authenticate(login, password)
}

// Login checks the credentials and logs the session in. See Manager.Login.
func Login(s session.SessionInterface, w http.ResponseWriter, login string, password string) bool {
	return DefaultManager().Login(s, w, login, password)
}

// Logout logs the session out. See Manager.Logout.
func Logout(s session.SessionInterface, w http.ResponseWriter) {
	DefaultManager().Logout(s, w)
}

// Remember sets a remember-me cookie for the logged-in user.
// See Manager.Remember.
func Remember(s session.SessionInterface, r *http.Request, w http.ResponseWriter, rememberIP bool) {
	DefaultManager().Remember(s, r, w, rememberIP)
}

// Restore logs the session in from a remember-me cookie.
// See Manager.Restore.
func Restore(s session.SessionInterface, r *http.Request, w http.ResponseWriter) UserInterface {
	return DefaultManager().Restore(s, r, w)
}

// PasswordHash returns the stored form of a password. See
// Manager.PasswordHash.
func PasswordHash(password string, salt string) string {
	return DefaultManager().PasswordHash(password, salt)
}
