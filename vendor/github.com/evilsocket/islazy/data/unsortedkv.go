package data

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/evilsocket/islazy/fs"
)

// UnsortedKV is a thread safe and unsorted key-value
// storage with optional persistency on disk.
type UnsortedKV struct {
	sync.Mutex
	fileName string
	m        map[string]string
	policy   FlushPolicy
}

// NewUnsortedKV creates a new UnsortedKV with the given flush policy.
// If fileName already exists, it will be deserialized and loaded.
func NewUnsortedKV(fileName string, flushPolicy FlushPolicy) (*UnsortedKV, error) {
	ukv := &UnsortedKV{
		fileName: fileName,
		m:        make(map[string]string),
		policy:   flushPolicy,
	}

	if fileName != "" && fs.Exists(fileName) {
		raw, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, err
		}

		decoder := gob.NewDecoder(bytes.NewReader(raw))
		if err = decoder.Decode(&ukv.m); err != nil {
			return nil, err
		}
	}

	return ukv, nil
}

// NewDiskUnsortedKV returns an UnsortedKV that flushed data on disk
// every time it gets updated.
func NewDiskUnsortedKV(fileName string) (*UnsortedKV, error) {
	return NewUnsortedKV(fileName, FlushOnEdit)
}

// NewDiskUnsortedKVReader returns an UnsortedKV from disk as a reader
// but it doesn't flush any modifications on disk.
func NewDiskUnsortedKVReader(fileName string) (*UnsortedKV, error) {
	return NewUnsortedKV(fileName, FlushNone)
}

// NewMemUnsortedKV returns an UnsortedKV that only lives in
// memory and never persists on disk.
func NewMemUnsortedKV() (*UnsortedKV, error) {
	return NewUnsortedKV("", FlushNone)
}

// MarshalJSON is used to serialize the UnsortedKV data structure to
// JSON correctly.
func (u *UnsortedKV) MarshalJSON() ([]byte, error) {
	u.Lock()
	defer u.Unlock()
	return json.Marshal(u.m)
}

// Has return true if name exists in the store.
func (u *UnsortedKV) Has(name string) bool {
	u.Lock()
	defer u.Unlock()
	_, found := u.m[name]
	return found
}

// Get return the value of the named object if present, or returns
// found as false otherwise.
func (u *UnsortedKV) Get(name string) (v string, found bool) {
	u.Lock()
	defer u.Unlock()
	v, found = u.m[name]
	return
}

// GetOr will return the value of the named object if present,
// or a default value.
func (u *UnsortedKV) GetOr(name, or string) string {
	if v, found := u.Get(name); found {
		return v
	}
	return or
}

func (u *UnsortedKV) flushUnlocked() error {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(u.m); err != nil {
		return err
	}
	return ioutil.WriteFile(u.fileName, buf.Bytes(), os.ModePerm)
}

// Flush flushes the store to disk if the flush policy
// is different than FlushNone
func (u *UnsortedKV) Flush() error {
	u.Lock()
	defer u.Unlock()
	if u.policy != FlushNone {
		return u.flushUnlocked()
	}
	return nil
}

func (u *UnsortedKV) onEdit() error {
	if u.policy == FlushOnEdit {
		return u.flushUnlocked()
	}
	return nil
}

// Set sets a value for a named object.
func (u *UnsortedKV) Set(name, value string) error {
	u.Lock()
	defer u.Unlock()
	u.m[name] = value
	return u.onEdit()
}

// Del deletes a named object from the store.
func (u *UnsortedKV) Del(name string) error {
	u.Lock()
	defer u.Unlock()
	delete(u.m, name)
	return u.onEdit()
}

// Clear deletes every named object from the store.
func (u *UnsortedKV) Clear() error {
	u.Lock()
	defer u.Unlock()
	u.m = make(map[string]string)
	return u.onEdit()
}

// Each iterates each named object in the store by
// executing the callback cb on them, if the callback
// returns true the iteration is interrupted.
func (u *UnsortedKV) Each(cb func(k, v string) bool) {
	u.Lock()
	defer u.Unlock()
	for k, v := range u.m {
		if stop := cb(k, v); stop {
			return
		}
	}
}

// Empty returns bool if the store is empty.
func (u *UnsortedKV) Empty() bool {
	u.Lock()
	defer u.Unlock()
	return len(u.m) == 0
}
