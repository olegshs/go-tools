package auth

// UserInterface is the interface that application user types implement.
// Password returns the stored password hash, as produced by
// Manager.PasswordHash, and Salt returns the user's own salt.
type UserInterface interface {
	Id() int64
	Login() string
	Password() string
	Salt() string
	IsAdmin() bool
}
