// Пакет file реализует драйвер для хранения данных в файлах.
package file

import (
	_ "crypto/md5"
	"encoding/binary"
	"fmt"
	"hash"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/olegshs/go-tools/cache/storage"
	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/helpers"
)

const (
	fileExtension = ".tmp"
)

type Storage struct {
	dir  string
	hash hash.Hash

	ttlDefault time.Duration
	ttlMax     time.Duration

	gc *helpers.Interval

	hits   int64
	misses int64
}

func NewStorage(conf Config) *Storage {
	stor := new(Storage)

	stor.dir = config.AbsPath(conf.Dir)

	h := helpers.HashByName(conf.Hash)
	if !h.Available() {
		panic("hash function is not available: " + conf.Hash)
	}
	stor.hash = h.New()

	stor.ttlDefault = conf.TTL.Default
	stor.ttlMax = conf.TTL.Maximum

	if conf.GC.Interval > 0 {
		stor.gc = helpers.NewInterval(conf.GC.Interval, stor.gcRun)
		stor.gc.Start()
	}

	return stor
}

func (stor *Storage) Get(key string) ([]byte, error) {
	p := stor.pathByKey(key)

	f, err := os.Open(p)
	if err != nil {
		atomic.AddInt64(&stor.misses, 1)

		if os.IsNotExist(err) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	defer f.Close()

	var exp int64

	err = binary.Read(f, binary.LittleEndian, &exp)
	if err != nil {
		atomic.AddInt64(&stor.misses, 1)
		return nil, err
	}

	if exp > 0 && time.Unix(exp, 0).Before(time.Now()) {
		atomic.AddInt64(&stor.misses, 1)
		return nil, storage.ErrExpired
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		atomic.AddInt64(&stor.misses, 1)
		return nil, err
	}

	atomic.AddInt64(&stor.hits, 1)
	return data, nil
}

func (stor *Storage) Set(key string, data []byte, ttl time.Duration) error {
	p := stor.pathByKey(key)

	d := path.Dir(p)
	if _, err := os.Stat(d); err != nil {
		err = os.MkdirAll(d, 0700)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	exp := storage.Expire(ttl, stor.ttlDefault, stor.ttlMax)

	err = binary.Write(f, binary.LittleEndian, exp)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (stor *Storage) Delete(key string) error {
	p := stor.pathByKey(key)

	err := os.Remove(p)
	if err != nil {
		return err
	}

	return nil
}

func (stor *Storage) DeleteAll() error {
	err := filepath.Walk(stor.dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if p[len(p)-len(fileExtension):] != fileExtension {
			return nil
		}

		err = os.Remove(p)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (stor *Storage) Hits() int64 {
	return stor.hits
}

func (stor *Storage) Misses() int64 {
	return stor.misses
}

func (stor *Storage) gcRun() {
	filepath.Walk(stor.dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if p[len(p)-len(fileExtension):] != fileExtension {
			return nil
		}

		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()

		var exp int64

		err = binary.Read(f, binary.LittleEndian, &exp)
		if err != nil {
			return nil
		}

		if exp <= 0 {
			return nil
		}

		if time.Unix(exp, 0).Before(time.Now()) {
			os.Remove(p)
		}

		return nil
	})
}

func (stor *Storage) pathByKey(key string) string {
	stor.hash.Reset()
	stor.hash.Write([]byte(key))
	h := stor.hash.Sum(nil)

	return fmt.Sprintf("%s/%02x/%x%s", stor.dir, h[0], h[1:], fileExtension)
}
