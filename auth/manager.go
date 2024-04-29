package auth

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/olegshs/go-tools/auth/token"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/helpers/typeconv"
	"github.com/olegshs/go-tools/session"
)

type Manager struct {
	conf           Config
	userFactory    UserFactoryInterface
	sessionManager session.ManagerInterface
}

func NewManager(conf Config, userFactory UserFactoryInterface, sessionManager session.ManagerInterface) ManagerInterface {
	mgr := new(Manager)
	mgr.conf = conf
	mgr.userFactory = userFactory
	mgr.sessionManager = sessionManager

	return mgr
}

func (mgr *Manager) UserId(s session.SessionInterface) int64 {
	return typeconv.Int64(s.Get(mgr.conf.Session.UserIdKey))
}

func (mgr *Manager) SetUserId(s session.SessionInterface, id int64) {
	s.Set(mgr.conf.Session.UserIdKey, id)
}

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

func (mgr *Manager) Authenticate(login string, password string) bool {
	user := mgr.userFactory.UserByLogin(login)
	if user == nil {
		return false
	}

	userPassword := user.Password()
	if len(userPassword) == 0 {
		return false
	}

	if mgr.PasswordHash(password, user.Salt()) != userPassword {
		return false
	}

	return true
}

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

func (mgr *Manager) Logout(s session.SessionInterface, w http.ResponseWriter) {
	mgr.sessionManager.Restart(s, w)

	c := new(http.Cookie)
	c.Name = mgr.conf.Cookie.Name
	c.Value = ""
	c.Path = "/"
	c.Domain = mgr.conf.Cookie.Domain
	c.Secure = mgr.conf.Cookie.Secure
	c.Expires = time.Now().Add(-time.Hour * 24)
	c.HttpOnly = true
	http.SetCookie(w, c)
}

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
