package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
// Credits to https://golangcode.com/unzip-files-in-go/
func Unzip(src string, dest string) ([]string, error) {
	var outFile *os.File
	var zipFile io.ReadCloser
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	clean := func() {
		if outFile != nil {
			outFile.Close()
			outFile = nil
		}

		if zipFile != nil {
			zipFile.Close()
			zipFile = nil
		}
	}

	for _, f := range r.File {
		zipFile, err = f.Open()
		if err != nil {
			return filenames, err
		}

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: https://snyk.io/research/zip-slip-vulnerability#go
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			clean()
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			clean()
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			clean()
			return filenames, err
		}

		outFile, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			clean()
			return filenames, err
		}

		_, err = io.Copy(outFile, zipFile)
		clean()
		if err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}
