// Пакет logger содержит общие для различных драйверов функции и определения.
package logger

import (
	"io"
)

// Интерфейс драйвера журнала.
type Logger interface {
	io.Writer
	Print(level int, a ...interface{}) error
}
