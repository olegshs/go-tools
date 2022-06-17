package interfaces

type Row interface {
	Scan(dest ...interface{}) error
}
