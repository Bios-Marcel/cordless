package files

import (
	"encoding/json"
	"io/ioutil"
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

func GetAbsolutePath(directoryPath string) (string, error) {
	absolutePath, resolveError := ToAbsolutePath(directoryPath)
	if resolveError != nil {
		return "", resolveError
	}
	return absolutePath, resolveError
}

func EnsureDirectory(directoryPath string) error {
	_, statError := os.Stat(directoryPath)
	if os.IsNotExist(statError) {
		createDirsError := os.MkdirAll(directoryPath, 0766)
		if createDirsError != nil {
			return createDirsError
		}
		return nil
	}
	return statError
}

func CheckExists(path string) error {
	_, statError := os.Stat(path)
	return statError
}

func LoadJSON(path string, store interface{}) error {
	existsError := CheckExists(path)
	if existsError != nil {
		return existsError
	}

	file, readError := ioutil.ReadFile(path)
	if readError != nil {
		return readError
	}

	jsonError := json.Unmarshal([]byte(file), &store)
	if jsonError != nil {
		return jsonError
	}

	return nil
}

func WriteJSON(path string, store interface{}) error {
	jsonContents, jsonError := json.MarshalIndent(store, "", "    ")
	if jsonError != nil {
		return jsonError
	}

	writeError := ioutil.WriteFile(path, jsonContents, 0644)
	if writeError != nil {
		return writeError
	}

	return nil

}
