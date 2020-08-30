package files

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// ToAbsolutePath handles different kind of paths and makes an absolute path
// out of them. Consider the following three inputs:
// 	   file:///home/marcel/test.txt%C3%A4
//     ./test.txtä
//     ~/test.txtä
// Those will be turned into  (assuming that our current working directory
// is /home/marcel:
//     /home/marcel/test.txtä
// However, this method doesn't check for file existence.
func ToAbsolutePath(input string) (string, error) {
	var resolvedPath string
	if strings.HasPrefix(input, "~") {
		currentUser, userResolveError := user.Current()
		if userResolveError != nil {
			return "", userResolveError
		}

		resolvedPath = filepath.Join(currentUser.HomeDir, strings.TrimPrefix(input, "~"))
	} else {
		resolvedPath = strings.TrimPrefix(input, "file://")
		var unescapeError error
		resolvedPath, unescapeError = url.PathUnescape(resolvedPath)
		if unescapeError != nil {
			return "", unescapeError
		}
	}

	resolvedPath, resolveError := filepath.Abs(resolvedPath)
	if resolveError != nil {
		return "", resolveError
	}

	return resolvedPath, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath, url string) error {
	response, httpError := http.Get(url)
	if httpError != nil {
		return httpError
	}
	defer response.Body.Close()

	outputFile, fileError := os.Create(filepath)
	if fileError != nil {
		return fileError
	}
	defer outputFile.Close()

	_, writeError := io.Copy(outputFile, response.Body)
	return writeError
}

// DownloadFileOrAccessCache checks whether the file already exists on the
// users filesystem and only downloads it if it doesn't.
func DownloadFileOrAccessCache(filepath, url string) error {
	_, statError := os.Stat(filepath)

	//File already exists
	if statError == nil {
		return nil
	}

	return DownloadFile(filepath, url)
}

// EnsureDirectoryExists creates a directy if doesn't already exist.
// If an error occurs during the existence check, that error is returned
// directly.
func EnsureDirectoryExists(directoryPath string) error {
	_, statError := os.Stat(directoryPath)
	if os.IsNotExist(statError) {
		return os.MkdirAll(directoryPath, 0766)
	}
	return statError
}
