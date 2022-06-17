// Пакет session предоставляет средства для работы с HTTP сессиями.
package session

import (
	"bytes"
	"encoding/gob"

	"github.com/olegshs/go-tools/events"
	"github.com/olegshs/go-tools/session/storage"
)

type Session struct {
	storage StorageInterface
	events  *events.Dispatcher

	id        string
	data      map[string]interface{}
	isChanged bool
}

func (s *Session) Id() string {
	return s.id
}

func (s *Session) SetId(id string) {
	s.id = id
	s.isChanged = true
}

func (s *Session) Get(key string) interface{} {
	return s.data[key]
}

func (s *Session) Set(key string, value interface{}) {
	s.data[key] = value
	s.isChanged = true
}

func (s *Session) Delete(key string) {
	delete(s.data, key)
	s.isChanged = true
}

func (s *Session) DeleteAll() {
	s.data = map[string]interface{}{}
	s.isChanged = true
}

func (s *Session) Destroy() error {
	if s.id == "" {
		return nil
	}

	return s.storage.Delete(s.id)
}

func (s *Session) Close() error {
	if s.id == "" {
		return nil
	}

	var err error

	if s.isChanged {
		err = s.save()
	} else {
		err = s.touch()
	}

	return err
}

func (s *Session) load() {
	if s.id == "" {
		return
	}

	s.data, _ = s.loadData()
	s.events.Dispatch(EventAfterLoad, s)
}

func (s *Session) loadData() (map[string]interface{}, error) {
	m := map[string]interface{}{}

	data, err := s.storage.Get(s.id)
	if err != nil && err != storage.ErrExpired {
		return m, err
	}

	if len(data) > 0 {
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(&m)
	}

	return m, err
}

func (s *Session) save() error {
	if s.id == "" {
		return nil
	}

	s.events.Dispatch(EventBeforeSave, s)

	buf := new(bytes.Buffer)

	if len(s.data) > 0 {
		enc := gob.NewEncoder(buf)
		err := enc.Encode(s.data)
		if err != nil {
			return err
		}
	}

	err := s.storage.Set(s.id, buf.Bytes())
	if err != nil {
		return err
	}

	s.isChanged = false

	return nil
}

func (s *Session) touch() error {
	if s.id == "" {
		return nil
	}

	toucher, ok := s.storage.(StorageTouchInterface)
	if !ok {
		err := s.save()
		if err != nil {
			return err
		}
		return nil
	}

	err := toucher.Touch(s.id)
	if err != nil {
		return err
	}

	return nil
}
