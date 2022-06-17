package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/events"
	"github.com/olegshs/go-tools/helpers"
)

const (
	StorageFile = "file"
)

type FileStorage struct {
	dir    string
	ttl    time.Duration
	gc     *helpers.Interval
	events *events.Dispatcher
}

func NewFileStorage(conf ConfigFile) *FileStorage {
	storage := new(FileStorage)

	storage.dir = config.AbsPath(conf.Dir)

	gcInterval := conf.GC.Interval
	if gcInterval > 0 {
		storage.gc = helpers.NewInterval(gcInterval, storage.gcRun)
		storage.gc.RandomDelay(gcInterval / 8)
		storage.gc.Start()
	}

	storage.events = events.New()

	return storage
}

func (stor *FileStorage) Events() *events.Dispatcher {
	return stor.events
}

func (stor *FileStorage) SetTTL(ttl time.Duration) {
	stor.ttl = ttl
}

func (stor *FileStorage) IsExist(id string) (bool, error) {
	p := stor.PathById(id)

	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (stor *FileStorage) Get(id string) ([]byte, error) {
	p := stor.PathById(id)

	fi, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	if stor.ttl > 0 && stor.IsExpired(fi.ModTime()) {
		return nil, ErrExpired
	}

	stor.events.Dispatch(EventAfterLoad, id, data)

	return data, nil
}

func (stor *FileStorage) Set(id string, data []byte) error {
	stor.events.Dispatch(EventBeforeSave, id, data)

	p := stor.PathById(id)

	if len(data) == 0 {
		_, err := os.Stat(p)
		if err == nil {
			err = os.Remove(p)
			if err != nil {
				return err
			}
		}

		return nil
	}

	d := path.Dir(p)
	if _, err := os.Stat(d); os.IsNotExist(err) {
		err := os.MkdirAll(d, 0700)
		if err != nil {
			return err
		}
	}

	err := ioutil.WriteFile(p, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (stor *FileStorage) Touch(id string) error {
	stor.events.Dispatch(EventBeforeTouch, id)

	p := stor.PathById(id)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	}

	t := time.Now()
	err := os.Chtimes(p, t, t)
	if err != nil {
		return err
	}

	return nil
}

func (stor *FileStorage) Delete(id string) error {
	isExist, err := stor.IsExist(id)
	if err != nil {
		return err
	}
	if !isExist {
		return ErrNotFound
	}

	stor.events.Dispatch(EventBeforeDelete, id)

	p := stor.PathById(id)

	err = os.Remove(p)
	if err != nil {
		return err
	}

	return nil
}

func (stor *FileStorage) Expire() time.Time {
	return time.Now().Add(-stor.ttl)
}

func (stor *FileStorage) IsExpired(t time.Time) bool {
	return t.Before(stor.Expire())
}

func (stor *FileStorage) PathById(id string) string {
	return fmt.Sprintf("%s/%s/%s", stor.dir, id[0:2], id[2:])
}

func (stor *FileStorage) gcRun() {
	if stor.ttl <= 0 {
		return
	}

	filepath.Walk(stor.dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if stor.IsExpired(info.ModTime()) {
			p := strings.Replace(p, "\\", "/", -1)
			a := strings.Split(p, "/")
			n := len(a)
			id := a[n-2] + a[n-1]

			stor.Delete(id)
		}

		return nil
	})
}
