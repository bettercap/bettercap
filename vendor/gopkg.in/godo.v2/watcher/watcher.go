// Package watcher implements filesystem notification,.
package watcher

import (
	//"fmt"

	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mgutz/str"
	"gopkg.in/godo.v2/watcher/fswatch"
)

const (
	// IgnoreThresholdRange is the amount of time in ns to ignore when
	// receiving watch events for the same file
	IgnoreThresholdRange = 50 * 1000000 // convert to ms
)

// SetWatchDelay sets the watch delay
func SetWatchDelay(delay time.Duration) {
	fswatch.WatchDelay = delay
}

// Watcher is a wrapper around which adds some additional features:
//
// - recursive directory watch
// - buffer to even chan
// - even time
//
// Original work from https://github.com/bronze1man/kmg
type Watcher struct {
	*fswatch.Watcher
	Event chan *FileEvent
	Error chan error
	//default ignore all file start with "."
	IgnorePathFn func(path string) bool
	//default is nil,if is nil ,error send through Error chan,if is not nil,error handle by this func
	ErrorHandler func(err error)
	isClosed     bool
	quit         chan bool
	cache        map[string]*os.FileInfo
	mu           sync.Mutex
}

// NewWatcher creates an instance of watcher.
func NewWatcher(bufferSize int) (watcher *Watcher, err error) {

	fswatcher := fswatch.NewAutoWatcher()

	if err != nil {
		return nil, err
	}
	watcher = &Watcher{
		Watcher:      fswatcher,
		Error:        make(chan error, 10),
		Event:        make(chan *FileEvent, bufferSize),
		IgnorePathFn: DefaultIgnorePathFn,
		cache:        map[string]*os.FileInfo{},
	}
	return
}

// Close closes the watcher channels.
func (w *Watcher) Close() error {
	if w.isClosed {
		return nil
	}
	w.Watcher.Stop()
	w.quit <- true
	w.isClosed = true
	return nil
}

func (w *Watcher) eventLoop() {
	// cache := map[string]*os.FileInfo{}
	// mu := &sync.Mutex{}

	coutput := w.Watcher.Start()
	for {
		event, ok := <-coutput
		if !ok {
			return
		}

		// fmt.Printf("event %+v\n", event)
		if w.IgnorePathFn(event.Path) {
			continue
		}

		// you can not stat a delete file...
		if event.Event == fswatch.DELETED || event.Event == fswatch.NOEXIST {
			// adjust with arbitrary value because it was deleted
			// before it got here
			//fmt.Println("sending fi wevent", event)
			w.Event <- newFileEvent(event.Event, event.Path, time.Now().UnixNano()-10)
			continue
		}

		fi, err := os.Stat(event.Path)
		if os.IsNotExist(err) {
			//fmt.Println("not exists", event)
			continue
		}

		// fsnotify is sending multiple MODIFY events for the same
		// file which is likely OS related. The solution here is to
		// compare the current stats of a file against its last stats
		// (if any) and if it falls within a nanoseconds threshold,
		// ignore it.
		w.mu.Lock()
		oldFI := w.cache[event.Path]
		w.cache[event.Path] = &fi

		// if oldFI != nil {
		// 	fmt.Println("new FI", fi.ModTime().UnixNano())
		// 	fmt.Println("old FI", (*oldFI).ModTime().UnixNano()+IgnoreThresholdRange)
		// }

		if oldFI != nil && fi.ModTime().UnixNano() < (*oldFI).ModTime().UnixNano()+IgnoreThresholdRange {
			w.mu.Unlock()
			continue
		}
		w.mu.Unlock()

		//fmt.Println("sending Event", fi.Name())

		//fmt.Println("sending fi wevent", event)
		w.Event <- newFileEvent(event.Event, event.Path, fi.ModTime().UnixNano())

		if err != nil {
			//rename send two events,one old file,one new file,here ignore old one
			if os.IsNotExist(err) {
				continue
			}
			w.errorHandle(err)
			continue
		}
		// case err := <-w.Watcher.Errors:
		// 	w.errorHandle(err)
		// case _ = <-w.quit:
		// 	break
		// }
	}
}
func (w *Watcher) errorHandle(err error) {
	if w.ErrorHandler == nil {
		w.Error <- err
		return
	}
	w.ErrorHandler(err)
}

// GetErrorChan gets error chan.
func (w *Watcher) GetErrorChan() chan error {
	return w.Error
}

// GetEventChan gets event chan.
func (w *Watcher) GetEventChan() chan *FileEvent {
	return w.Event
}

// WatchRecursive watches a directory recursively. If a dir is created
// within directory it is also watched.
func (w *Watcher) WatchRecursive(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	w.Watcher.Add(path)

	//util.Debug("watcher", "watching %s %s\n", path, time.Now())
	return nil
}

// Start starts the watcher
func (w *Watcher) Start() {
	go w.eventLoop()
}

// func (w *Watcher) getSubFolders(path string) (paths []string, err error) {
// 	err = filepath.Walk(path, func(newPath string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		if !info.IsDir() {
// 			return nil
// 		}
// 		if w.IgnorePathFn(newPath) {
// 			return filepath.SkipDir
// 		}
// 		paths = append(paths, newPath)
// 		return nil
// 	})
// 	return paths, err
// }

// DefaultIgnorePathFn checks whether a path is ignored. Currently defaults
// to hidden files on *nix systems, ie they start with a ".".
func DefaultIgnorePathFn(path string) bool {
	if strings.HasPrefix(path, ".") || strings.Contains(path, "/.") {
		return true
	}

	// ignore node
	if strings.HasPrefix(path, "node_modules") || strings.Contains(path, "/node_modules") {
		return true
	}

	// vim creates random numeric files
	base := filepath.Base(path)
	if str.IsNumeric(base) {
		return true
	}
	return false
}

// SetIgnorePathFn sets the function which determines if a path should be
// skipped when watching.
func (w *Watcher) SetIgnorePathFn(fn func(string) bool) {
	w.Watcher.IgnorePathFn = fn
}
