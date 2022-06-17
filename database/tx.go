package database

import (
	"database/sql"
	"time"

	"github.com/olegshs/go-tools/database/interfaces"
	"github.com/olegshs/go-tools/database/query"
	"github.com/olegshs/go-tools/events"
)

type Tx struct {
	db *DB
	tx *sql.Tx
}

func (tx *Tx) Driver() string {
	return tx.db.Driver()
}

func (tx *Tx) DB() *sql.DB {
	return tx.db.DB()
}

func (tx *Tx) Helper() interfaces.Helper {
	return tx.db.Helper()
}

func (tx *Tx) Events() *events.Dispatcher {
	return tx.db.Events()
}

func (tx *Tx) Commit() error {
	err := tx.tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

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

func (tx *Tx) QueryRow(query string, args ...interface{}) interfaces.Row {
	t0 := time.Now()
	rows, err := tx.tx.Query(query, args...)
	t1 := time.Now()

	tx.db.events.Dispatch(EventQueryRow, t0, t1, query, args, err)

	row := &Row{rows, err}
	return row
}

func (tx *Tx) Rollback() error {
	err := tx.tx.Rollback()
	if err != nil {
		return err
	}

	return nil
}

func (tx *Tx) Select(columns ...interface{}) interfaces.Query {
	q := query.New(tx, tx.db.helper)
	q.Select(columns...)
	return q
}

func (tx *Tx) Insert(table string, data interface{}) interfaces.Query {
	q := query.New(tx, tx.db.helper)
	q.Insert(table, data)
	return q
}

func (tx *Tx) Update(table string, data interface{}) interfaces.Query {
	q := query.New(tx, tx.db.helper)
	q.Update(table, data)
	return q
}

func (tx *Tx) Delete(table string) interfaces.Query {
	q := query.New(tx, tx.db.helper)
	q.Delete(table)
	return q
}
