package fs

import (
	"path/filepath"
)

// Glob enumerates files on a given path using a globbing expression and
// execute a callback for each of the files. The callback can interrupt
// the loop by returning an error other than nil.
func Glob(path string, expr string, cb func(fileName string) error) (err error) {
	var files []string
	if files, err = filepath.Glob(filepath.Join(path, expr)); err == nil {
		for _, fileName := range files {
			if err = cb(fileName); err != nil {
				return
			}
		}
	}
	return
}
