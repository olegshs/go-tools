// Package mysql connects to MySQL. It requires the database/sql driver
// registered as "mysql", such as github.com/go-sql-driver/mysql.
package mysql

import (
	"database/sql"
	"fmt"
)

// New opens a connection pool and checks it with a ping.
func New(conf Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s%s",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.Database, conf.Params,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
