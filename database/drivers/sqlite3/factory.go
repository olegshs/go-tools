// Package sqlite3 connects to SQLite. It requires the database/sql driver
// registered as "sqlite3", such as github.com/mattn/go-sqlite3.
package sqlite3

import (
	"database/sql"
	"fmt"
)

// New opens the database file and checks it with a ping.
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
