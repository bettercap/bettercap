package api_rest

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/evilsocket/islazy/fs"
	"github.com/kr/binarydist"
)

type patch []byte
type frame []byte

type RecordEntry struct {
	sync.Mutex

	Data      []byte  `json:"data"`
	Cur       []byte  `json:"-"`
	States    []patch `json:"states"`
	NumStates int     `json:"-"`
	CurState  int     `json:"-"`

	frames []frame
}

func NewRecordEntry() *RecordEntry {
	return &RecordEntry{
		Data:      nil,
		Cur:       nil,
		States:    make([]patch, 0),
		NumStates: 0,
		CurState:  0,
		frames:    nil,
	}
}

func (e *RecordEntry) AddState(state []byte) error {
	e.Lock()
	defer e.Unlock()

	// set reference state
	if e.Data == nil {
		e.Data = state
	} else {
		// create a patch
		oldReader := bytes.NewReader(e.Cur)
		newReader := bytes.NewReader(state)
		writer := new(bytes.Buffer)

		if err := binarydist.Diff(oldReader, newReader, writer); err != nil {
			return err
		}

		e.States = append(e.States, patch(writer.Bytes()))
		e.NumStates++
		e.CurState = 0
	}
	e.Cur = state

	return nil
}

func (e *RecordEntry) Reset() {
	e.Lock()
	defer e.Unlock()
	e.Cur = e.Data
	e.NumStates = len(e.States)
	e.CurState = 0
}

func (e *RecordEntry) Compile() error {
	e.Lock()
	defer e.Unlock()

	// reset the state
	e.Cur = e.Data
	e.NumStates = len(e.States)
	e.CurState = 0
	e.frames = make([]frame, e.NumStates+1)

	// first is the master frame
	e.frames[0] = frame(e.Data)
	// precompute frames so they can be accessed by index
	for i := 0; i < e.NumStates; i++ {
		patch := e.States[i]
		oldReader := bytes.NewReader(e.Cur)
		patchReader := bytes.NewReader(patch)
		newWriter := new(bytes.Buffer)

		if err := binarydist.Patch(oldReader, newWriter, patchReader); err != nil {
			return err
		}

		e.Cur = newWriter.Bytes()
		e.frames[i+1] = e.Cur
	}

	return nil
}

func (e *RecordEntry) Frames() int {
	e.Lock()
	defer e.Unlock()
	// master + sub states
	return e.NumStates + 1
}

func (e *RecordEntry) CurFrame() int {
	e.Lock()
	defer e.Unlock()
	return e.CurState + 1
}

func (e *RecordEntry) SetFrom(from int) {
	e.Lock()
	defer e.Unlock()
	e.CurState = from
}

func (e *RecordEntry) Over() bool {
	e.Lock()
	defer e.Unlock()
	return e.CurState > e.NumStates
}

func (e *RecordEntry) Next() []byte {
	e.Lock()
	defer e.Unlock()
	cur := e.CurState
	e.CurState++
	return e.frames[cur]
}

// the Record object represents a recorded session
type Record struct {
	sync.Mutex

	fileName string       `json:"-"`
	Session  *RecordEntry `json:"session"`
	Events   *RecordEntry `json:"events"`
}

func NewRecord(fileName string) *Record {
	return &Record{
		fileName: fileName,
		Session:  NewRecordEntry(),
		Events:   NewRecordEntry(),
	}
}

func LoadRecord(fileName string) (*Record, error) {
	if !fs.Exists(fileName) {
		return nil, fmt.Errorf("%s does not exist", fileName)
	}

	compressed, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("error while reading %s: %s", fileName, err)
	}

	decompress, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("error while reading gzip file %s: %s", fileName, err)
	}
	defer decompress.Close()

	raw, err := ioutil.ReadAll(decompress)
	if err != nil {
		return nil, fmt.Errorf("error while decompressing %s: %s", fileName, err)
	}

	rec := &Record{}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err = decoder.Decode(rec); err != nil {
		return nil, fmt.Errorf("error while parsing %s: %s", fileName, err)
	}

	rec.fileName = fileName

	// reset state and precompute frames
	if err = rec.Session.Compile(); err != nil {
		return nil, err
	} else if err = rec.Events.Compile(); err != nil {
		return nil, err
	}

	return rec, nil
}

func (r *Record) NewState(session []byte, events []byte) error {
	if err := r.Session.AddState(session); err != nil {
		return err
	} else if err := r.Events.AddState(events); err != nil {
		return err
	}
	return r.Flush()
}

func (r *Record) save() error {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)

	if err := encoder.Encode(r); err != nil {
		return err
	}

	data := buf.Bytes()

	compressed := new(bytes.Buffer)
	compress := gzip.NewWriter(compressed)

	if _, err := compress.Write(data); err != nil {
		return err
	} else if err = compress.Flush(); err != nil {
		return err
	} else if err = compress.Close(); err != nil {
		return err
	}

	return ioutil.WriteFile(r.fileName, compressed.Bytes(), os.ModePerm)
}

func (r *Record) Flush() error {
	r.Lock()
	defer r.Unlock()
	return r.save()
}
