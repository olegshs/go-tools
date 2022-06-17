// Пакет interfaces определяет общие ошибки для различных пакетов.
package errors

import (
	"errors"
)

var (
	UnknownDriver    = errors.New("unknown driver")
	UnknownDatabase  = errors.New("unknown database")
	InvalidStatement = errors.New("invalid statement")
)
