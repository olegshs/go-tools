package interfaces

import (
	"database/sql"
)

// Result is the result of a statement, as returned by database/sql.
type Result interface {
	sql.Result
}
