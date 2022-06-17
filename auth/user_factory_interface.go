package auth

type UserFactoryInterface interface {
	UserById(int64) UserInterface
	UserByLogin(string) UserInterface
}
