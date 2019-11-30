package config

import (
	"os/user"
	"path/filepath"
)

func getDefaultConfigDirectory() (string, error) {
	// TODO Gotta research this

	currentUser, userError := user.Current()

	if userError != nil {
		return "", userError
	}

	return filepath.Join(currentUser.HomeDir, "."+AppNameLowercase), nil
}
