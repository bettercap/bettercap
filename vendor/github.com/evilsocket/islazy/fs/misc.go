package fs

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Expand will expand a path with ~ to a full path of the current user.
func Expand(path string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Abs(strings.Replace(path, "~", usr.HomeDir, 1))
}

// Exists returns true if the path exists.
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
