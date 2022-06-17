package storage

import (
	"database/sql"
	"time"

	"github.com/olegshs/go-tools/database"
	"github.com/olegshs/go-tools/database/query"
	"github.com/olegshs/go-tools/events"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/logs"
	"github.com/olegshs/go-tools/session/storage/encoder"
)

const (
	StorageSql = "sql"
)

type SqlStorage struct {
	db      string
	table   string
	encoder encoder.Encoder
	ttl     time.Duration
	gc      *helpers.Interval
	events  *events.Dispatcher
}

func NewSqlStorage(conf ConfigSql) *SqlStorage {
	storage := new(SqlStorage)
	storage.db = conf.Database
	storage.table = conf.Table

	encoding := conf.Encoding
	storage.encoder = encoder.New(encoding)

	gcInterval := conf.GC.Interval
	if gcInterval > 0 {
		storage.gc = helpers.NewInterval(gcInterval, storage.gcRun)
		storage.gc.Start()
	}

	storage.events = events.New()

	return storage
}

func (stor *SqlStorage) Events() *events.Dispatcher {
	return stor.events
}

func (stor *SqlStorage) SetTTL(ttl time.Duration) {
	stor.ttl = ttl
}

func (stor *SqlStorage) IsExist(id string) (bool, error) {
	var count int

	db, err := database.Get(stor.db)
	if err != nil {
		logs.Error("SqlStorage.IsExist/DB:", err)
		return false, err
	}

	row := db.Select(
		query.Expr("COUNT(*)"),
	).From(
		stor.table,
	).Where(
		query.Eq{
			"id": id,
		},
	).Row()

	err = row.Scan(&count)
	if err != nil {
		logs.Error("SqlStorage.IsExist:", err)
		return false, err
	}

	return count > 0, nil
}

func (stor *SqlStorage) Get(id string) ([]byte, error) {
	var (
		t    int64
		data []byte
	)

	db, err := database.Get(stor.db)
	if err != nil {
		logs.Error("SqlStorage.Get/DB:", err)
		return nil, err
	}

	row := db.Select(
		"time",
		"data",
	).From(
		stor.table,
	).Where(
		query.Eq{
			"id": id,
		},
	).Row()

	err = row.Scan(&t, &data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		logs.Error("SqlStorage.Get:", err)
		return nil, err
	}

	data, err = stor.encoder.Decode(data)
	if err != nil {
		return nil, err
	}

	if stor.ttl > 0 && stor.IsExpired(time.Unix(t/1000, t%1000*1000000)) {
		return nil, ErrExpired
	}

	stor.events.Dispatch(EventAfterLoad, id, data)

	return data, nil
}

func (stor *SqlStorage) Set(id string, data []byte) error {
	t := time.Now().UnixNano() / 1000000

	stor.events.Dispatch(EventBeforeSave, id, data)

	db, err := database.Get(stor.db)
	if err != nil {
		logs.Error("SqlStorage.Set/DB:", err)
		return err
	}

	if len(data) == 0 {
		_, err := db.Delete(
			stor.table,
		).Where(
			query.Eq{
				"id": id,
			},
		).Exec()

		if err != nil {
			logs.Error("SqlStorage.Set/Delete:", err)
			return err
		}
		return nil
	}

	data, err = stor.encoder.Encode(data)
	if err != nil {
		return err
	}

	res, err := db.Update(
		stor.table,
		query.Data{
			"time": t,
			"data": data,
		},
	).Where(
		query.Eq{
			"id": id,
		},
	).Exec()

	if err != nil {
		logs.Error("SqlStorage.Set/Update:", err)
		return err
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		_, err := db.Insert(
			stor.table,
			query.Data{
				"id":   id,
				"time": t,
				"data": data,
			},
		).Exec()

		if err != nil {
			logs.Error("SqlStorage.Set/Insert:", err)
			return err
		}
	}

	return nil
}

func (stor *SqlStorage) Touch(id string) error {
	t := time.Now().UnixNano() / 1000000

	stor.events.Dispatch(EventBeforeTouch, id)

	db, err := database.Get(stor.db)
	if err != nil {
		logs.Error("SqlStorage.Touch/DB:", err)
		return err
	}

	_, err = db.Update(
		stor.table,
		query.Data{
			"time": t,
		},
	).Where(
		query.Eq{
			"id": id,
		},
	).Exec()

	if err != nil {
		logs.Error("SqlStorage.Touch:", err)
		return err
	}

	return nil
}

func (stor *SqlStorage) Delete(id string) error {
	stor.events.Dispatch(EventBeforeDelete, id)

	db, err := database.Get(stor.db)
	if err != nil {
		logs.Error("SqlStorage.Delete/DB:", err)
		return err
	}

	_, err = db.Delete(
		stor.table,
	).Where(
		query.Eq{
			"id": id,
		},
	).Exec()

	if err != nil {
		logs.Error("SqlStorage.Delete:", err)
		return err
	}

	return nil
}

func (stor *SqlStorage) Expire() time.Time {
	return time.Now().Add(-stor.ttl)
}

func (stor *SqlStorage) IsExpired(t time.Time) bool {
	return t.Before(stor.Expire())
}

func (stor *SqlStorage) gcRun() {
	if stor.ttl <= 0 {
		return
	}

	exp := stor.Expire().UnixNano() / 1000000

	db, err := database.Get(stor.db)
	if err != nil {
		logs.Error("SqlStorage.gcRun/DB:", err)
		return
	}

	rows, err := db.Select(
		"id",
	).From(
		stor.table,
	).Where(
		query.Lte{
			"time": exp,
		},
	).Rows()
	if err != nil {
		logs.Error("SqlStorage.gcRun:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string

		err := rows.Scan(&id)
		if err != nil {
			continue
		}

		stor.Delete(id)
	}
}
