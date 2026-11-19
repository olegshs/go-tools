package auth

// UserFactoryInterface loads users for authentication. Both methods return
// nil if the user does not exist.
type UserFactoryInterface interface {
	UserById(int64) UserInterface
	UserByLogin(string) UserInterface
}
