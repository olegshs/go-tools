// Пакет mysql реализует драйвер для работы с MySQL.
package mysql

import (
	"database/sql"
	"fmt"
)

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
