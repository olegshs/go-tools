package auth

type dummyUserFactory struct {
}

func (f *dummyUserFactory) UserById(int64) UserInterface {
	return nil
}

func (f *dummyUserFactory) UserByLogin(string) UserInterface {
	return nil
}
