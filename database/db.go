// Пакет database предоставляет обёртку для работы с базами данных.
package database

import (
	"database/sql"
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
	dbInstances      = map[string]*DB{}
	dbInstancesMutex sync.Mutex
)

const (
	DefaultDB = "default"

	DriverMysql    = "mysql"
	DriverPostgres = "postgres"
	DriverSqlite3  = "sqlite3"
)

type DB struct {
	driver string
	db     *sql.DB
	helper interfaces.Helper
	events *events.Dispatcher
}

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

func (db *DB) Driver() string {
	return db.driver
}

func (db *DB) DB() *sql.DB {
	return db.db
}

func (db *DB) Helper() interfaces.Helper {
	return db.helper
}

func (db *DB) Events() *events.Dispatcher {
	return db.events
}

func (db *DB) Begin() (*Tx, error) {
	t, err := db.db.Begin()
	if err != nil {
		return nil, err
	}

	tx := &Tx{db, t}
	return tx, nil
}

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

func (db *DB) QueryRow(query string, args ...interface{}) interfaces.Row {
	t0 := time.Now()
	rows, err := db.db.Query(query, args...)
	t1 := time.Now()

	db.events.Dispatch(EventQueryRow, t0, t1, query, args, err)

	row := &Row{rows, err}
	return row
}

func (db *DB) Select(columns ...interface{}) interfaces.Query {
	q := query.New(db, db.helper)
	q.Select(columns...)
	return q
}

func (db *DB) Insert(table string, data interface{}) interfaces.Query {
	q := query.New(db, db.helper)
	q.Insert(table, data)
	return q
}

func (db *DB) Update(table string, data interface{}) interfaces.Query {
	q := query.New(db, db.helper)
	q.Update(table, data)
	return q
}

func (db *DB) Delete(table string) interfaces.Query {
	q := query.New(db, db.helper)
	q.Delete(table)
	return q
}

func (db *DB) Transaction(f func(*Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = f(tx)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		return nil
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
