package session

import (
	"time"

	"github.com/olegshs/go-tools/events"
)

type StorageInterface interface {
	Events() *events.Dispatcher
	SetTTL(time.Duration)
	IsExist(string) (bool, error)
	Get(string) ([]byte, error)
	Set(string, []byte) error
	Delete(string) error
}
