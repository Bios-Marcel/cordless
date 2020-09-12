package files

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AddToZip writes the given data to a zip archive using the passed writer.
// Folders are added to the zip recursively. The folder structure is fully
// preserved.
func AddToZip(zipWriter *zip.Writer, filename string) error {
	info, statError := os.Stat(filename)
	if os.IsNotExist(statError) || statError != nil {
		return statError
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(filename)
	}

	filepath.Walk(filename, func(path string, info os.FileInfo, readError error) error {
		if readError != nil {
			return readError
		}

		header, readError := zip.FileInfoHeader(info)
		if readError != nil {
			return readError
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, filename))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, headerError := zipWriter.CreateHeader(header)
		if headerError != nil {
			return headerError
		}

		if info.IsDir() {
			return nil
		}

		file, readError := os.Open(path)
		if readError != nil {
			return readError
		}

		defer file.Close()
		_, readError = io.Copy(writer, file)
		return readError
	})

	return nil
}
