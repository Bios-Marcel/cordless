package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	//AppName is the name representing the application.
	AppName = "Cordless"
	//AppNameLowercase is the representative name, but lowercase.
	//It us used for filepaths and such.
	AppNameLowercase = "cordless"

	HourMinuteAndSeconds = 0
	HourAndMinute        = 1
	NoTime               = 2
)

var (
	currentConfig = Config{}
)

//Config contains all possible configuration for the application.
type Config struct {
	Token string

	Times             int
	ShowUserContainer bool
	UseFrame          bool
}

var cachedConfigDir string

//GetConfigFile returns the absolute path to the configuration file or an error
//in case of failure.
func GetConfigFile() (string, error) {
	configDir, configError := GetConfigDirectory()

	if configError != nil {
		return "", configError
	}

	return filepath.Join(configDir, "config.json"), nil
}

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

	_, statError := os.Stat(directory)
	if os.IsNotExist(statError) {
		createDirsError := os.MkdirAll(directory, 0755)
		if createDirsError != nil {
			return "", createDirsError
		}
	} else if statError != nil {
		return "", statError
	}

	//After first retrieval, we will save this, as we needn't redo all that
	//stuff over and over again.
	cachedConfigDir = directory

	return cachedConfigDir, nil
}

//GetConfig returns the currently loaded configuration.
func GetConfig() *Config {
	return &currentConfig
}

//LoadConfig loads the configuration initially and returns it.
func LoadConfig() (*Config, error) {
	configFilePath, configError := GetConfigFile()
	if configError != nil {
		return nil, configError
	}

	configFile, openError := os.Open(configFilePath)

	if os.IsNotExist(openError) {
		return GetConfig(), nil
	}

	if openError != nil {
		return nil, openError
	}

	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	configLoadError := decoder.Decode(&currentConfig)

	//io.EOF would mean empty, therefore we use defaults.
	if configLoadError != nil && configLoadError != io.EOF {
		return nil, configLoadError
	}

	return GetConfig(), nil
}

//PersistConfig saves the current configuration onto the filesystem.
func PersistConfig() error {
	configFilePath, configError := GetConfigFile()
	if configError != nil {
		return configError
	}

	configAsJSON, jsonError := json.MarshalIndent(&currentConfig, "", "    ")
	if jsonError != nil {
		return jsonError
	}

	writeError := ioutil.WriteFile(configFilePath, configAsJSON, 0755)
	if writeError != nil {
		return writeError
	}

	return nil
}
