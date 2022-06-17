package database

import (
	"database/sql"
)

type Row struct {
	rows *sql.Rows
	err  error
}

func (r *Row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}

	defer r.rows.Close()

	if !r.rows.Next() {
		err := r.rows.Err()
		if err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	err := r.rows.Scan(dest...)
	if err != nil {
		return err
	}

	err = r.rows.Close()
	if err != nil {
		return err
	}

	return nil
}
