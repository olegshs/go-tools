// Пакет sqlite3 реализует драйвер для работы с SQLite.
package sqlite3

import (
	"database/sql"
	"fmt"
)

func New(conf Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"file:%s%s",
		conf.File, conf.Params,
	)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
