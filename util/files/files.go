package files

import (
	"net/url"
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
