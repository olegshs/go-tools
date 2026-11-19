package interfaces

// Row is the result of a single-row query.
type Row interface {
	Scan(dest ...interface{}) error
}
