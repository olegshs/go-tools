package database

import (
	"database/sql"
)

// Row is the result of QueryRow. It implements interfaces.Row.
type Row struct {
	rows *sql.Rows
	err  error
}

// Scan copies the columns of the first row into dest and closes the rows.
// It returns the query error, if any, or sql.ErrNoRows if there is no row.
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
