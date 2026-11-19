package session

import (
	"time"

	"github.com/olegshs/go-tools/events"
)

// StorageInterface is implemented by session storages. Get returns
// storage.ErrNotFound or storage.ErrExpired when there is no usable data,
// and Set with empty data deletes the session.
type StorageInterface interface {
	Events() *events.Dispatcher
	SetTTL(time.Duration)
	IsExist(string) (bool, error)
	Get(string) ([]byte, error)
	Set(string, []byte) error
	Delete(string) error
}
