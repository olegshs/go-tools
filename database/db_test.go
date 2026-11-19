package database

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
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
	defer os.Remove(f.Name())
	defer f.Close()

	db, err := Get(DefaultDB)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer db.Close()

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

	err = updatePostsWithStmt(db)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = testTransaction(db)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = dropTablePosts(db)
	if err != nil {
		t.Fatal(err)
		return
	}

	for _, err := range eventTests.wait(time.Second) {
		t.Errorf("event test failed: %s", err)
	}
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

// eventTests records whether each query event was dispatched with the
// expected arguments. Events are dispatched asynchronously, so the results
// are guarded by a mutex.
type eventTests struct {
	mutex   sync.Mutex
	results map[events.Event]error
}

func initEventTests(db *DB) *eventTests {
	errNotDispatched := errors.New("not dispatched")

	et := &eventTests{
		results: map[events.Event]error{
			EventExec:     errNotDispatched,
			EventPrepare:  errNotDispatched,
			EventQuery:    errNotDispatched,
			EventQueryRow: errNotDispatched,
		},
	}

	for event := range et.results {
		db.events.AddListener(event, et.callback(db, event))
	}

	return et
}

func (et *eventTests) callback(db *DB, event events.Event) events.Callback {
	var f func(args ...interface{})
	f = func(args ...interface{}) {
		var err error

		if len(args) != 5 {
			err = errors.New("invalid number of arguments")
		}

		et.mutex.Lock()
		et.results[event] = err
		et.mutex.Unlock()

		if err != nil {
			db.events.RemoveListener(event, f)
		}
	}
	return f
}

// wait returns the failed event tests, waiting up to timeout for listeners
// that have not run yet.
func (et *eventTests) wait(timeout time.Duration) []error {
	deadline := time.Now().Add(timeout)

	for {
		var errs []error

		et.mutex.Lock()
		for event, err := range et.results {
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", event, err))
			}
		}
		et.mutex.Unlock()

		if len(errs) == 0 || time.Now().After(deadline) {
			return errs
		}

		time.Sleep(10 * time.Millisecond)
	}
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
	defer stmt.Close()

	rows, err := stmt.Query(1)
	if err != nil {
		return err
	}

	// Open rows keep the SQLite database locked.
	err = rows.Close()
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

func updatePostsWithStmt(db *DB) error {
	stmt, err := db.Prepare(`UPDATE "posts" SET "status" = $1 WHERE "name" = $2`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(2, "hello")
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("stmt.Exec: %d rows affected, expected 1", n)
	}

	return nil
}

func testTransaction(db *DB) error {
	errAbort := errors.New("abort")

	err := db.Transaction(func(tx *Tx) error {
		_, err := tx.Insert("posts", query.Data{"name": "rolled back"}).Exec()
		if err != nil {
			return err
		}
		return errAbort
	})
	if !errors.Is(err, errAbort) {
		return fmt.Errorf("Transaction returned %v, expected %v", err, errAbort)
	}

	err = db.Transaction(func(tx *Tx) error {
		_, err := tx.Insert("posts", query.Data{"name": "committed"}).Exec()
		return err
	})
	if err != nil {
		return err
	}

	for name, expected := range map[string]int{"rolled back": 0, "committed": 1} {
		var count int

		err = db.Select(query.Expr("COUNT(*)")).
			From("posts").
			Where(query.Eq{"name": name}).
			Row().
			Scan(&count)
		if err != nil {
			return err
		}

		if count != expected {
			return fmt.Errorf("Transaction: %d %q rows, expected %d", count, name, expected)
		}
	}

	return nil
}

func dropTablePosts(db *DB) error {
	_, err := db.Exec(`DROP TABLE "posts"`)
	return err
}
