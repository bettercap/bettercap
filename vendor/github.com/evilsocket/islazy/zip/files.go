package zip

import (
	"archive/zip"
	"io"
	"os"
)

// Files compresses one or many files into a single zip archive file.
// Credits: https://golangcode.com/create-zip-files-in-go/
func Files(filename string, files []string) error {
	arc, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer arc.Close()

	writer := zip.NewWriter(arc)
	defer writer.Close()

	for _, file := range files {
		in, err := os.Open(file)
		if err != nil {
			return err
		}
		defer in.Close()

		info, err := in.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Using FileInfoHeader() above only uses the basename of the file. If we want
		// to preserve the folder structure we can overwrite this with the full path.
		header.Name = file
		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		w, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(w, in); err != nil {
			return err
		}
	}

	return nil
}
