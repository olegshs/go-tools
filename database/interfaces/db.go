package interfaces

import (
	"database/sql"

	"github.com/olegshs/go-tools/events"
)

type DB interface {
	Driver() string
	DB() *sql.DB
	Helper() Helper
	Events() *events.Dispatcher
	Exec(query string, args ...interface{}) (Result, error)
	Prepare(query string) (Stmt, error)
	Query(query string, args ...interface{}) (Rows, error)
	QueryRow(query string, args ...interface{}) Row
	Select(columns ...interface{}) Query
	Insert(table string, data interface{}) Query
	Update(table string, data interface{}) Query
	Delete(table string) Query
}
