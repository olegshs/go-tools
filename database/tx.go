package database

import (
	"database/sql"
	"time"

	"github.com/olegshs/go-tools/database/interfaces"
	"github.com/olegshs/go-tools/database/query"
	"github.com/olegshs/go-tools/events"
)

// Tx is a transaction. It implements interfaces.DB, so the query builder
// and the ORM can run inside it, and it dispatches events on the events of
// its DB.
type Tx struct {
	db *DB
	tx *sql.Tx
}

// Driver returns the driver name.
func (tx *Tx) Driver() string {
	return tx.db.Driver()
}

// DB returns the underlying *sql.DB of the connection, not the transaction.
func (tx *Tx) DB() *sql.DB {
	return tx.db.DB()
}

// Helper returns the SQL dialect helper of the driver.
func (tx *Tx) Helper() interfaces.Helper {
	return tx.db.Helper()
}

// Events returns the dispatcher of the connection.
func (tx *Tx) Events() *events.Dispatcher {
	return tx.db.Events()
}

// Commit commits the transaction.
func (tx *Tx) Commit() error {
	err := tx.tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Exec runs a statement in the transaction. See DB.Exec.
func (tx *Tx) Exec(query string, args ...interface{}) (interfaces.Result, error) {
	t0 := time.Now()
	res, err := tx.tx.Exec(query, args...)
	t1 := time.Now()

	tx.db.events.Dispatch(EventExec, t0, t1, query, args, err)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Prepare creates a prepared statement in the transaction.
func (tx *Tx) Prepare(query string) (interfaces.Stmt, error) {
	t0 := time.Now()
	s, err := tx.tx.Prepare(query)
	t1 := time.Now()

	tx.db.events.Dispatch(EventPrepare, t0, t1, query, nil, err)

	if err != nil {
		return nil, err
	}

	stmt := &Stmt{tx.db, s, query}
	return stmt, nil
}

// Query runs a query in the transaction. See DB.Query.
func (tx *Tx) Query(query string, args ...interface{}) (interfaces.Rows, error) {
	t0 := time.Now()
	rows, err := tx.tx.Query(query, args...)
	t1 := time.Now()

	tx.db.events.Dispatch(EventQuery, t0, t1, query, args, err)

	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRow runs a single-row query in the transaction. See DB.QueryRow.
func (tx *Tx) QueryRow(query string, args ...interface{}) interfaces.Row {
	t0 := time.Now()
	rows, err := tx.tx.Query(query, args...)
	t1 := time.Now()

	tx.db.events.Dispatch(EventQueryRow, t0, t1, query, args, err)

	row := &Row{rows, err}
	return row
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() error {
	err := tx.tx.Rollback()
	if err != nil {
		return err
	}

	return nil
}

// Select starts a SELECT query in the transaction.
func (tx *Tx) Select(columns ...interface{}) interfaces.Query {
	q := query.New(tx, tx.db.helper)
	q.Select(columns...)
	return q
}

// Insert starts an INSERT query in the transaction.
func (tx *Tx) Insert(table string, data interface{}) interfaces.Query {
	q := query.New(tx, tx.db.helper)
	q.Insert(table, data)
	return q
}

// Update starts an UPDATE query in the transaction.
func (tx *Tx) Update(table string, data interface{}) interfaces.Query {
	q := query.New(tx, tx.db.helper)
	q.Update(table, data)
	return q
}

// Delete starts a DELETE query in the transaction.
func (tx *Tx) Delete(table string) interfaces.Query {
	q := query.New(tx, tx.db.helper)
	q.Delete(table)
	return q
}
