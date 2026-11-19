// Package session implements server-side HTTP sessions identified by a
// random ID in a cookie.
//
// Session data is a map of values encoded with encoding/gob, so custom types
// stored in a session must be registered with gob.Register. Data is kept in
// file or SQL storage (see package session/storage), chosen by the
// "session.storage.driver" config setting.
//
// A typical handler starts the session, uses it, and closes it to save
// changes:
//
//	s := session.Start(r, w)
//	defer s.Close()
//	n, _ := s.Get("visits").(int)
//	s.Set("visits", n+1)
//
// Start and Restart may set the session cookie, so call them before writing
// the response. Close only writes to storage and can be deferred.
package session

import (
	"bytes"
	"encoding/gob"

	"github.com/olegshs/go-tools/events"
	"github.com/olegshs/go-tools/session/storage"
)

// Session holds the data of one session. It is not safe for concurrent
// use; open a separate Session in each request.
type Session struct {
	storage StorageInterface
	events  *events.Dispatcher

	id   string
	data map[string]interface{}
	// isChanged is set when data or the ID changes, so Close knows whether to
	// save the data or only refresh its expiry time.
	isChanged bool
}

// Id returns the session ID.
func (s *Session) Id() string {
	return s.id
}

// SetId changes the session ID. The data is saved under the new ID on
// Close. It does not change the cookie; see Manager.Restart.
func (s *Session) SetId(id string) {
	s.id = id
	s.isChanged = true
}

// Get returns the value of key, or nil if it is not set.
func (s *Session) Get(key string) interface{} {
	return s.data[key]
}

// Set stores value under key. The value's type must be encodable by gob.
func (s *Session) Set(key string, value interface{}) {
	s.data[key] = value
	s.isChanged = true
}

// Delete removes key from the session.
func (s *Session) Delete(key string) {
	delete(s.data, key)
	s.isChanged = true
}

// DeleteAll removes all values from the session.
func (s *Session) DeleteAll() {
	s.data = map[string]interface{}{}
	s.isChanged = true
}

// Destroy deletes the session from storage. The in-memory data is kept, so
// a later Close with pending changes stores it again.
func (s *Session) Destroy() error {
	if s.id == "" {
		return nil
	}

	return s.storage.Delete(s.id)
}

// Close saves the session if it changed. Otherwise it only refreshes the
// session's expiry time in storage.
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

// load reads the session data from storage and dispatches EventAfterLoad.
// Missing, expired or undecodable data gives an empty session.
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

// save dispatches EventBeforeSave and writes the data to storage. An empty
// session is written as empty data, which storages treat as a delete.
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

// touch refreshes the expiry time of the session. Storages that do not
// implement StorageTouchInterface get a full save instead.
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
