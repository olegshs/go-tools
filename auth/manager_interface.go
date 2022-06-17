package auth

import (
	"net/http"

	"github.com/olegshs/go-tools/session"
)

type ManagerInterface interface {
	UserId(s session.SessionInterface) int64
	SetUserId(s session.SessionInterface, id int64)
	User(s session.SessionInterface) UserInterface
	Authenticate(login string, password string) bool
	Login(s session.SessionInterface, w http.ResponseWriter, login string, password string) bool
	Logout(s session.SessionInterface, w http.ResponseWriter)
	Remember(s session.SessionInterface, r *http.Request, w http.ResponseWriter, rememberIP bool)
	Restore(s session.SessionInterface, r *http.Request, w http.ResponseWriter) UserInterface
	PasswordHash(password string, salt string) string
}
