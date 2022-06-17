// Пакет postgres реализует драйвер для работы с PostgreSQL.
package postgres

import (
	"database/sql"
	"fmt"
)

func New(conf Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s%s",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.Database, conf.Params,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
