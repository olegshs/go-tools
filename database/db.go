// Package database wraps database/sql with named connections from config,
// a fluent query builder and query events.
//
// Connections are configured under "database.<name>":
//
//	database:
//	  default:
//	    driver: postgres   # mysql, postgres or sqlite3
//	    host: localhost
//	    database: app
//	    log:
//	      enabled: true
//
// The application must import the database/sql driver itself, for example
// _ "github.com/lib/pq", _ "github.com/go-sql-driver/mysql" or
// _ "github.com/mattn/go-sqlite3".
//
// Every query dispatches an event with its SQL, arguments, timing and error
// (see EventExec and the others), which the optional query log uses. The
// query builder is in package database/query, and a reflection-based ORM is
// in package database/orm.
package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/olegshs/go-tools/config"
	dbConfig "github.com/olegshs/go-tools/database/config"
	"github.com/olegshs/go-tools/database/drivers/mysql"
	"github.com/olegshs/go-tools/database/drivers/postgres"
	"github.com/olegshs/go-tools/database/drivers/sqlite3"
	"github.com/olegshs/go-tools/database/errors"
	"github.com/olegshs/go-tools/database/interfaces"
	"github.com/olegshs/go-tools/database/query"
	"github.com/olegshs/go-tools/events"
)

var (
	// dbInstances caches open connections by name.
	dbInstances      = map[string]*DB{}
	dbInstancesMutex sync.Mutex
)

const (
	// DefaultDB is the name of the connection used by the package-level
	// functions.
	DefaultDB = "default"

	// Driver names, used in the "driver" config setting.
	DriverMysql    = "mysql"
	DriverPostgres = "postgres"
	DriverSqlite3  = "sqlite3"
)

// DB is a named database connection pool. It implements interfaces.DB and
// is safe for concurrent use.
type DB struct {
	driver string
	db     *sql.DB
	helper interfaces.Helper
	events *events.Dispatcher
}

// Get returns the connection with the given name, opening it from config on
// first use. It returns errors.UnknownDatabase if there is no
// "database.<name>" section.
func Get(name string) (*DB, error) {
	dbInstancesMutex.Lock()
	defer dbInstancesMutex.Unlock()

	db, ok := dbInstances[name]
	if ok {
		return db, nil
	}

	db, err := New(name)
	if err != nil {
		return nil, err
	}

	dbInstances[name] = db

	return db, nil
}

// New opens a new connection from the "database.<name>" config section
// without caching it, and checks it with a ping. It starts the query log if
// it is enabled.
func New(name string) (*DB, error) {
	confKey := "database." + name
	if !config.Exists(confKey) {
		return nil, errors.UnknownDatabase
	}

	conf := dbConfig.DefaultConfig()
	config.GetStruct(confKey, &conf)

	sqlDB, helper, err := NewSqlDB(conf.Driver, confKey)
	if err != nil {
		return nil, err
	}

	db := new(DB)
	db.driver = conf.Driver
	db.db = sqlDB
	db.helper = helper
	db.events = events.New()

	if conf.Log.Enabled {
		log := NewLog(db, conf.Log)
		log.Start()
	}

	return db, nil
}

// NewSqlDB opens a *sql.DB for the driver with the settings at confKey, and
// returns it with the driver's SQL dialect helper. It returns
// errors.UnknownDriver for unknown drivers.
func NewSqlDB(driver, confKey string) (*sql.DB, interfaces.Helper, error) {
	var (
		sqlDB  *sql.DB
		helper interfaces.Helper
		err    error
	)

	switch driver {
	case DriverMysql:
		conf := mysql.DefaultConfig()
		config.GetStruct(confKey, &conf)

		sqlDB, err = mysql.New(conf)
		helper = new(mysql.Helper)

	case DriverPostgres:
		conf := postgres.DefaultConfig()
		config.GetStruct(confKey, &conf)

		sqlDB, err = postgres.New(conf)
		helper = new(postgres.Helper)

	case DriverSqlite3:
		conf := sqlite3.DefaultConfig()
		config.GetStruct(confKey, &conf)

		sqlDB, err = sqlite3.New(conf)
		helper = new(sqlite3.Helper)

	default:
		err = errors.UnknownDriver
	}

	if err != nil {
		return nil, nil, err
	}
	return sqlDB, helper, nil
}

// Driver returns the driver name.
func (db *DB) Driver() string {
	return db.driver
}

// DB returns the underlying *sql.DB. Queries run on it directly do not
// dispatch events.
func (db *DB) DB() *sql.DB {
	return db.db
}

// Helper returns the SQL dialect helper of the driver.
func (db *DB) Helper() interfaces.Helper {
	return db.helper
}

// Events returns the dispatcher for query events of this connection and
// its transactions.
func (db *DB) Events() *events.Dispatcher {
	return db.events
}

// Begin starts a transaction.
func (db *DB) Begin() (*Tx, error) {
	t, err := db.db.Begin()
	if err != nil {
		return nil, err
	}

	tx := &Tx{db, t}
	return tx, nil
}

// Close closes the connection pool and removes it from the cache, so the
// next Get opens a new one.
func (db *DB) Close() error {
	err := db.db.Close()
	if err != nil {
		return err
	}

	dbInstancesMutex.Lock()
	defer dbInstancesMutex.Unlock()

	for k, v := range dbInstances {
		if v == db {
			delete(dbInstances, k)
			break
		}
	}

	return nil
}

// Ping checks the connection and dispatches EventPing.
func (db *DB) Ping() error {
	t0 := time.Now()
	err := db.db.Ping()
	t1 := time.Now()

	db.events.Dispatch(EventPing, t0, t1, nil, nil, err)

	if err != nil {
		return err
	}
	return nil
}

// Exec runs a statement that returns no rows and dispatches EventExec.
func (db *DB) Exec(query string, args ...interface{}) (interfaces.Result, error) {
	t0 := time.Now()
	res, err := db.db.Exec(query, args...)
	t1 := time.Now()

	db.events.Dispatch(EventExec, t0, t1, query, args, err)

	if err != nil {
		return nil, err
	}
	return res, nil
}

// Prepare creates a prepared statement and dispatches EventPrepare.
func (db *DB) Prepare(query string) (interfaces.Stmt, error) {
	t0 := time.Now()
	s, err := db.db.Prepare(query)
	t1 := time.Now()

	db.events.Dispatch(EventPrepare, t0, t1, query, nil, err)

	if err != nil {
		return nil, err
	}

	stmt := &Stmt{db, s, query}
	return stmt, nil
}

// Query runs a query and dispatches EventQuery. The caller must close the
// returned rows.
func (db *DB) Query(query string, args ...interface{}) (interfaces.Rows, error) {
	t0 := time.Now()
	rows, err := db.db.Query(query, args...)
	t1 := time.Now()

	db.events.Dispatch(EventQuery, t0, t1, query, args, err)

	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRow runs a query that is expected to return one row and dispatches
// EventQueryRow. Errors are returned by the row's Scan. The row holds a
// connection until Scan is called, so Scan must always be called.
func (db *DB) QueryRow(query string, args ...interface{}) interfaces.Row {
	t0 := time.Now()
	rows, err := db.db.Query(query, args...)
	t1 := time.Now()

	db.events.Dispatch(EventQueryRow, t0, t1, query, args, err)

	row := &Row{rows, err}
	return row
}

// Select starts a SELECT query on this connection. See package query.
func (db *DB) Select(columns ...interface{}) interfaces.Query {
	q := query.New(db, db.helper)
	q.Select(columns...)
	return q
}

// Insert starts an INSERT query on this connection. data is a query.Data
// map or a SELECT *query.Query.
func (db *DB) Insert(table string, data interface{}) interfaces.Query {
	q := query.New(db, db.helper)
	q.Insert(table, data)
	return q
}

// Update starts an UPDATE query on this connection. data is a query.Data
// map.
func (db *DB) Update(table string, data interface{}) interfaces.Query {
	q := query.New(db, db.helper)
	q.Update(table, data)
	return q
}

// Delete starts a DELETE query on this connection.
func (db *DB) Delete(table string) interfaces.Query {
	q := query.New(db, db.helper)
	q.Delete(table)
	return q
}

// Transaction calls f inside a transaction. The transaction is committed if
// f returns nil and rolled back otherwise. It returns the error from f
// (wrapped together with the rollback error if the rollback fails too) or
// the commit error.
func (db *DB) Transaction(f func(*Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = f(tx)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("%w (rollback failed: %w)", err, rollbackErr)
		}
		return err
	}

	return tx.Commit()
}
