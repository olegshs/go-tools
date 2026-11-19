package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/olegshs/go-tools/auth/token"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/helpers/typeconv"
	"github.com/olegshs/go-tools/session"
)

// Manager implements authentication with a user factory and a session
// manager.
type Manager struct {
	conf           Config
	userFactory    UserFactoryInterface
	sessionManager session.ManagerInterface
}

// NewManager creates an auth manager.
func NewManager(conf Config, userFactory UserFactoryInterface, sessionManager session.ManagerInterface) ManagerInterface {
	mgr := new(Manager)
	mgr.conf = conf
	mgr.userFactory = userFactory
	mgr.sessionManager = sessionManager

	return mgr
}

// UserId returns the ID of the user logged in to the session, or 0.
func (mgr *Manager) UserId(s session.SessionInterface) int64 {
	return typeconv.Int64(s.Get(mgr.conf.Session.UserIdKey))
}

// SetUserId stores the user ID in the session. It does not restart the
// session; Login does that.
func (mgr *Manager) SetUserId(s session.SessionInterface, id int64) {
	s.Set(mgr.conf.Session.UserIdKey, id)
}

// User loads the user logged in to the session, or returns nil.
func (mgr *Manager) User(s session.SessionInterface) UserInterface {
	id := mgr.UserId(s)
	if id <= 0 {
		return nil
	}

	user := mgr.userFactory.UserById(id)
	if user == nil {
		return nil
	}

	return user
}

// Authenticate reports whether login names a user whose stored password
// hash matches password. Users with an empty stored password cannot log in.
func (mgr *Manager) Authenticate(login string, password string) bool {
	user := mgr.userFactory.UserByLogin(login)
	if user == nil {
		return false
	}

	userPassword := user.Password()
	if len(userPassword) == 0 {
		return false
	}

	// Compare in constant time, so the response time does not reveal how
	// much of the hash matched.
	hash := mgr.PasswordHash(password, user.Salt())
	if subtle.ConstantTimeCompare([]byte(hash), []byte(userPassword)) != 1 {
		return false
	}

	return true
}

// Login checks the credentials and, if they are valid, restarts the session
// with a new ID and stores the user ID in it. It reports whether the login
// succeeded.
func (mgr *Manager) Login(s session.SessionInterface, w http.ResponseWriter, login string, password string) bool {
	if !mgr.Authenticate(login, password) {
		return false
	}

	user := mgr.userFactory.UserByLogin(login)
	if user == nil {
		return false
	}

	mgr.sessionManager.Restart(s, w)
	mgr.SetUserId(s, user.Id())

	return true
}

// Logout clears the session, gives it a new ID and deletes the remember-me
// cookie.
func (mgr *Manager) Logout(s session.SessionInterface, w http.ResponseWriter) {
	mgr.sessionManager.Restart(s, w)

	c := new(http.Cookie)
	c.Name = mgr.conf.Cookie.Name
	c.Value = ""
	c.Path = mgr.conf.Cookie.Path
	c.Domain = mgr.conf.Cookie.Domain
	c.Secure = mgr.conf.Cookie.Secure
	c.Expires = time.Now().Add(-time.Hour * 24)
	c.HttpOnly = true
	http.SetCookie(w, c)
}

// Remember sets a signed remember-me cookie for the user logged in to the
// session. It does nothing if no user is logged in or the hash is not
// available.
//
// The signature covers the payload, the configured secret and the user's
// stored password hash, so changing the password invalidates existing
// cookies. With rememberIP, it also covers the client IP address.
func (mgr *Manager) Remember(s session.SessionInterface, r *http.Request, w http.ResponseWriter, rememberIP bool) {
	id := mgr.UserId(s)
	if id <= 0 {
		return
	}

	user := mgr.userFactory.UserById(id)
	if user == nil {
		return
	}

	t := time.Now()

	payload := new(TokenPayload)
	payload.UserId = id
	payload.Time = uint32(t.Unix())
	payload.RememberIP = rememberIP

	secrets := []interface{}{
		mgr.conf.Cookie.Secret,
		user.Password(),
	}
	if rememberIP {
		ip := helpers.RequestIP(r)
		secrets = append(secrets, ip)
	}

	h := helpers.HashByName(mgr.conf.Cookie.Hash)
	if !h.Available() {
		return
	}

	tokenBytes, err := token.Encode(payload, h, secrets...)
	if err != nil {
		return
	}

	tokenString := base64.RawURLEncoding.EncodeToString(tokenBytes)

	c := new(http.Cookie)
	c.Name = mgr.conf.Cookie.Name
	c.Value = tokenString
	c.Path = mgr.conf.Cookie.Path
	c.Domain = mgr.conf.Cookie.Domain
	c.Secure = mgr.conf.Cookie.Secure
	c.Expires = t.Add(mgr.conf.Cookie.TTL)
	c.HttpOnly = true
	http.SetCookie(w, c)
}

// Restore checks the remember-me cookie and, if it is valid and not older
// than the TTL, restarts the session with a new ID, logs the user in and
// returns the user. It returns nil if there is no valid cookie.
func (mgr *Manager) Restore(s session.SessionInterface, r *http.Request, w http.ResponseWriter) UserInterface {
	c, err := r.Cookie(mgr.conf.Cookie.Name)
	if err != nil {
		return nil
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(c.Value)
	if err != nil {
		return nil
	}

	payload := new(TokenPayload)
	err = token.Decode(payloadBytes, payload)
	if err != nil {
		return nil
	}

	id := payload.UserId
	if id <= 0 {
		return nil
	}

	t := time.Unix(int64(payload.Time), 0)
	exp := time.Now().Add(-mgr.conf.Cookie.TTL)
	if t.Before(exp) {
		return nil
	}

	user := mgr.userFactory.UserById(id)
	if user == nil {
		return nil
	}

	// These must match the secrets used by Remember.
	secrets := []interface{}{
		mgr.conf.Cookie.Secret,
		user.Password(),
	}
	if payload.RememberIP {
		ip := helpers.RequestIP(r)
		secrets = append(secrets, ip)
	}

	h := helpers.HashByName(mgr.conf.Cookie.Hash)
	if !h.Available() {
		return nil
	}

	err = token.Validate(payloadBytes, h, secrets...)
	if err != nil {
		return nil
	}

	mgr.sessionManager.Restart(s, w)
	mgr.SetUserId(s, user.Id())

	return user
}

// PasswordHash returns the hex digest of password + salt + the global salt
// with the configured hash, or password itself if the hash is "none". An
// empty password gives "". It panics if the hash is not available.
//
// This is a single fast hash, not a password hashing function such as
// bcrypt or Argon2, so stored hashes are cheap to brute-force if leaked.
func (mgr *Manager) PasswordHash(password string, salt string) string {
	if len(password) == 0 {
		return ""
	}

	hashName := mgr.conf.Password.Hash
	if hashName == "none" {
		return password
	}

	hash := helpers.HashByName(hashName)
	if !hash.Available() {
		panic("hash function is not available: " + hashName)
	}

	data := []byte(password + salt + mgr.conf.Password.Salt)

	h := hash.New()
	h.Write(data)
	b := h.Sum(nil)
	s := hex.EncodeToString(b[:])

	return s
}
