package database

import (
	"github.com/olegshs/go-tools/database/interfaces"
)

// Begin starts a transaction on the default connection.
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

// Ping checks the default connection.
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

// Exec runs a statement on the default connection. See DB.Exec.
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

// Prepare creates a prepared statement on the default connection.
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

// Query runs a query on the default connection. See DB.Query.
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

// QueryRow runs a single-row query on the default connection. See
// DB.QueryRow.
func QueryRow(query string, args ...interface{}) interfaces.Row {
	db, err := Get(DefaultDB)
	if err != nil {
		return &Row{nil, err}
	}

	row := db.QueryRow(query, args...)
	return row
}

// Select starts a SELECT query on the default connection. It returns nil if
// the connection cannot be opened.
func Select(columns ...interface{}) interfaces.Query {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil
	}

	return db.Select(columns...)
}

// Insert starts an INSERT query on the default connection. It returns nil
// if the connection cannot be opened.
func Insert(table string, data interface{}) interfaces.Query {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil
	}

	return db.Insert(table, data)
}

// Update starts an UPDATE query on the default connection. It returns nil
// if the connection cannot be opened.
func Update(table string, data interface{}) interfaces.Query {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil
	}

	return db.Update(table, data)
}

// Delete starts a DELETE query on the default connection. It returns nil if
// the connection cannot be opened.
func Delete(table string) interfaces.Query {
	db, err := Get(DefaultDB)
	if err != nil {
		return nil
	}

	return db.Delete(table)
}
