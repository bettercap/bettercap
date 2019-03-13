package fs

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
)

var (
	cwdLock = sync.Mutex{}
)

// Expand will expand a path with ~ to the $HOME of the current user.
func Expand(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	home := os.Getenv("HOME")
	if home == "" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		home = usr.HomeDir
	}
	return filepath.Abs(strings.Replace(path, "~", home, 1))
}

// Exists returns true if the path exists.
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// Chdir changes the process current working directory to the specified
// one, executes the callback and then restores the original working directory.
func Chdir(path string, cb func() error) error {
	cwdLock.Lock()
	defer cwdLock.Unlock()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	// make sure that whatever happens we restore the original
	// working directory of the process
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			panic(err)
		}
	}()
	// change folder
	if err := os.Chdir(path); err != nil {
		return err
	}
	// run the callback once inside the folder
	return cb()
}
