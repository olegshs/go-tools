package database

import (
	"database/sql"
	"time"

	"github.com/olegshs/go-tools/database/interfaces"
)

type Stmt struct {
	db    *DB
	stmt  *sql.Stmt
	query string
}

func (s *Stmt) Close() error {
	err := s.stmt.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *Stmt) Exec(args ...interface{}) (interfaces.Result, error) {
	t0 := time.Now()
	res, err := s.stmt.Exec(args)
	t1 := time.Now()

	s.db.events.Dispatch(EventExec, t0, t1, s.query, args, err)

	if err != nil {
		return nil, err
	}
	return res, nil
}

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

func (s *Stmt) QueryRow(args ...interface{}) interfaces.Row {
	t0 := time.Now()
	rows, err := s.stmt.Query(args...)
	t1 := time.Now()

	s.db.events.Dispatch(EventQueryRow, t0, t1, s.query, args, err)

	row := &Row{rows, err}
	return row
}
