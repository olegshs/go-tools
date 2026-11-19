package interfaces

// Stmt is a prepared statement.
type Stmt interface {
	Close() error
	Exec(args ...interface{}) (Result, error)
	Query(args ...interface{}) (Rows, error)
	QueryRow(args ...interface{}) Row
}
