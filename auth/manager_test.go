package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/olegshs/go-tools/session"
)

// stubSessionManager implements only the session manager methods used by
// the tests; calling any other method panics.
type stubSessionManager struct {
	session.ManagerInterface
}

func (stubSessionManager) Restart(session.SessionInterface, http.ResponseWriter) {}

type testUser struct {
	id       int64
	login    string
	password string
	salt     string
}

func (u *testUser) Id() int64        { return u.id }
func (u *testUser) Login() string    { return u.login }
func (u *testUser) Password() string { return u.password }
func (u *testUser) Salt() string     { return u.salt }
func (u *testUser) IsAdmin() bool    { return false }

type testUserFactory struct {
	user *testUser
}

func (f testUserFactory) UserById(id int64) UserInterface {
	if id != f.user.id {
		return nil
	}
	return f.user
}

func (f testUserFactory) UserByLogin(login string) UserInterface {
	if login != f.user.login {
		return nil
	}
	return f.user
}

func TestManager_Authenticate(t *testing.T) {
	user := &testUser{id: 1, login: "admin", salt: "pepper"}
	mgr := NewManager(DefaultConfig(), testUserFactory{user}, stubSessionManager{})
	user.password = mgr.PasswordHash("secret", user.salt)

	tests := []struct {
		login    string
		password string
		expected bool
	}{
		{"admin", "secret", true},
		{"admin", "wrong", false},
		{"admin", "", false},
		{"nobody", "secret", false},
	}

	for _, test := range tests {
		ok := mgr.Authenticate(test.login, test.password)
		if ok != test.expected {
			t.Errorf("Authenticate(%q, %q) = %v, expected %v", test.login, test.password, ok, test.expected)
		}
	}
}

func TestManager_LogoutCookiePath(t *testing.T) {
	conf := DefaultConfig()
	conf.Cookie.Path = "/app"
	mgr := NewManager(conf, new(dummyUserFactory), stubSessionManager{})

	w := httptest.NewRecorder()
	mgr.Logout(nil, w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookies set, expected 1", len(cookies))
	}

	c := cookies[0]
	if c.Name != conf.Cookie.Name || c.Value != "" || c.Path != conf.Cookie.Path {
		t.Errorf("cookie %q=%q with path %q, expected %q=\"\" with path %q",
			c.Name, c.Value, c.Path, conf.Cookie.Name, conf.Cookie.Path)
	}
}
