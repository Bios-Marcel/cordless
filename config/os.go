// +build !darwin,!windows

package config

import (
	"os"
	"os/user"
	"path/filepath"
)

func getDefaultConfigDirectory() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")

	if configDir != "" {
		return filepath.Join(configDir, AppNameLowercase), nil
	}

	currentUser, userError := user.Current()

	if userError != nil {
		return "", userError
	}

	return filepath.Join(currentUser.HomeDir, ".config", AppNameLowercase), nil
}
