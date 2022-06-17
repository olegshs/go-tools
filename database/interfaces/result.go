package interfaces

import (
	"database/sql"
)

type Result interface {
	sql.Result
}
