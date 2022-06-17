package database

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/database/query"
	"github.com/olegshs/go-tools/events"
)

var (
	tmpDir = "."
)

func init() {
	dir := os.TempDir()
	if _, err := os.Stat(dir); err == nil {
		tmpDir = dir
	}
}

func TestDB(t *testing.T) {
	f, err := initDatabase()
	if err != nil {
		t.Fatal(err)
		return
	}

	db, err := Get(DefaultDB)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = db.Ping()
	if err != nil {
		t.Fatal(err)
		return
	}

	eventTests := initEventTests(db)

	err = createTablePosts(db)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = insertIntoPosts(db)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = selectFromPosts(db)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = dropTablePosts(db)
	if err != nil {
		t.Fatal(err)
		return
	}

	for event, err := range eventTests {
		if err != nil {
			t.Errorf("event test failed: %s: %s", event, err)
		}
	}

	f.Close()
	os.Remove(f.Name())
}

func initDatabase() (*os.File, error) {
	f, err := ioutil.TempFile(tmpDir, "test.*.db")
	if err != nil {
		return nil, err
	}

	config.Set("database", map[string]interface{}{
		DefaultDB: map[string]interface{}{
			"driver": DriverSqlite3,
			"file":   f.Name(),
			"params": map[string]interface{}{},
		},
	})

	return f, nil
}

func initEventTests(db *DB) map[events.Event]error {
	errNotDispatched := errors.New("not dispatched")
	results := map[events.Event]error{
		EventExec:     errNotDispatched,
		EventPrepare:  errNotDispatched,
		EventQuery:    errNotDispatched,
		EventQueryRow: errNotDispatched,
	}
	callbacks := map[events.Event]events.Callback{}

	for event := range results {
		f := initEventCallback(db, results, event)

		callbacks[event] = f
		db.events.AddListener(event, f)
	}

	return results
}

func initEventCallback(db *DB, results map[events.Event]error, event events.Event) events.Callback {
	var f func(args ...interface{})
	f = func(args ...interface{}) {
		var err error

		if len(args) != 5 {
			err = errors.New("invalid number of arguments")
		}

		results[event] = err

		if err != nil {
			db.events.RemoveListener(event, f)
		}
	}
	return f
}

func createTablePosts(db *DB) error {
	_, err := db.Exec(`
		CREATE TABLE "posts" (
			"id"       INTEGER PRIMARY KEY AUTOINCREMENT,
			"user_id"  INTEGER,
			"name"     TEXT,
			"title"    TEXT,
			"content"  TEXT,
			"created"  INTEGER,
			"modified" INTEGER,
			"status"   INTEGER
		)
	`)
	return err
}

func insertIntoPosts(db *DB) error {
	now := time.Now()

	_, err := db.Insert("posts", query.Data{
		"user_id":  1,
		"name":     "hello",
		"title":    "Hello, world!",
		"content":  "Hello, world!  \nThis is my first post :)\n",
		"created":  now.Unix(),
		"modified": now.Unix(),
		"status":   1,
	}).Exec()

	return err
}

func selectFromPosts(db *DB) error {
	stmt, err := db.Prepare(db.
		Select("id", "name").
		From("posts").
		Where(query.Eq{
			"user_id": 0,
		}).
		String(),
	)
	if err != nil {
		return err
	}

	_, err = stmt.Query(1)
	if err != nil {
		return err
	}

	row := db.
		Select("id").
		From("posts").
		Where(query.Eq{
			"name": "hello",
		}).
		Row()

	var id int64
	err = row.Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

func dropTablePosts(db *DB) error {
	_, err := db.Exec(`DROP TABLE "posts"`)
	return err
}
