package session

// SessionInterface is the interface of a session. It is implemented by
// Session; see its methods for details.
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
