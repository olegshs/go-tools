package database

import (
	"database/sql"
	"time"

	"github.com/olegshs/go-tools/database/interfaces"
)

// Stmt is a prepared statement. It implements interfaces.Stmt and
// dispatches the same events as DB.
type Stmt struct {
	db    *DB
	stmt  *sql.Stmt
	query string
}

// Close closes the statement.
func (s *Stmt) Close() error {
	err := s.stmt.Close()
	if err != nil {
		return err
	}

	return nil
}

// Exec runs the statement with args.
func (s *Stmt) Exec(args ...interface{}) (interfaces.Result, error) {
	t0 := time.Now()
	res, err := s.stmt.Exec(args...)
	t1 := time.Now()

	s.db.events.Dispatch(EventExec, t0, t1, s.query, args, err)

	if err != nil {
		return nil, err
	}
	return res, nil
}

// Query runs the statement as a query. The caller must close the returned
// rows.
func (s *Stmt) Query(args ...interface{}) (interfaces.Rows, error) {
	t0 := time.Now()
	rows, err := s.stmt.Query(args...)
	t1 := time.Now()

	s.db.events.Dispatch(EventQuery, t0, t1, s.query, args, err)

	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRow runs the statement as a single-row query. See DB.QueryRow.
func (s *Stmt) QueryRow(args ...interface{}) interfaces.Row {
	t0 := time.Now()
	rows, err := s.stmt.Query(args...)
	t1 := time.Now()

	s.db.events.Dispatch(EventQueryRow, t0, t1, s.query, args, err)

	row := &Row{rows, err}
	return row
}
