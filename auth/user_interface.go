package auth

type UserInterface interface {
	Id() int64
	Login() string
	Password() string
	Salt() string
	IsAdmin() bool
}
