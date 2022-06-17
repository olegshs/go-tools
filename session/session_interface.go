package session

type SessionInterface interface {
	Id() string
	SetId(id string)
	Get(key string) interface{}
	Set(key string, value interface{})
	Delete(key string)
	DeleteAll()
	Destroy() error
	Close() error
}
