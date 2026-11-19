package interfaces

import (
	"database/sql"
)

// Rows is the result of a query. It has the methods of *sql.Rows.
type Rows interface {
	Next() bool
	NextResultSet() bool
	Err() error
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
	Scan(dest ...interface{}) error
	Close() error
}
