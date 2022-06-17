package database

import (
	"github.com/olegshs/go-tools/database/interfaces"
)

func Begin() (*Tx, error) {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func Ping() error {
	db, err := Get(DefaultDB)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func Exec(query string, args ...interface{}) (interfaces.Result, error) {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil, err
	}

	res, err := db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func Prepare(query string) (interfaces.Stmt, error) {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}

	return stmt, err
}

func Query(query string, args ...interface{}) (interfaces.Rows, error) {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func QueryRow(query string, args ...interface{}) interfaces.Row {
	db, err := Get(DefaultDB)
	if err != nil {
		return &Row{nil, err}
	}

	row := db.QueryRow(query, args...)
	return row
}

func Select(columns ...interface{}) interfaces.Query {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil
	}

	return db.Select(columns...)
}

func Insert(table string, data interface{}) interfaces.Query {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil
	}

	return db.Insert(table, data)
}

func Update(table string, data interface{}) interfaces.Query {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil
	}

	return db.Update(table, data)
}

func Delete(table string) interfaces.Query {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil
	}

	return db.Delete(table)
}
