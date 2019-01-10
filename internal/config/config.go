package config

import "strings"

var (
	//AppName is the name representing the application.
	AppName = "Cordless"
	//AppNameLowercase is the representative name, but lowercase.
	//It us used for filepaths and such.
	AppNameLowercase = strings.ToLower(AppName)
)

//Config contains all possible configuration for the application.
type Config struct {
	Token string
}

var cachedConfigDir string

//GetConfigDirectory is the parent directory in the os, that contains the
//settings for the application.
func GetConfigDirectory() (string, error) {
	if cachedConfigDir != "" {
		return cachedConfigDir, nil
	}

	directory, err := getConfigDirectory()
	if err != nil {
		return "", err
	}

	cachedConfigDir = directory
	return cachedConfigDir, nil
}
