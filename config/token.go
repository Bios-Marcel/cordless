package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

var Token string
var cachedTokenFile string

// UpdateCurrentToken updates the current token and all accounts where the
// token was also used.
func UpdateCurrentToken(newToken string) {
	oldToken := Token
	Token = newToken
	for _, account := range Accounts {
		if account.Token == oldToken {
			account.Token = newToken
		}
	}
}

// SetTokenFile sets the token file path cache to the
// entered value
func SetTokenFile(tokenFilePath string) error {
	// get parent directory of config file
	parent := filepath.Dir(tokenFilePath)
	err := ensureDirectory(parent)
	if err == nil {
		cachedTokenFile = tokenFilePath
	} else {
		absolute, err := getAbsolutePath(parent)
		if err != nil {
			return err
		}
		err = ensureDirectory(absolute)
		if err == nil {
			cachedTokenFile = tokenFilePath
		}
	}
	return err
}

// GetTokenFile retrieves the config file path from cache
// or sets it to the default config file location
func GetTokenFile() (string, error) {
	if cachedTokenFile != "" {
		return cachedTokenFile, nil
	}

	configDir, configDirError := GetConfigDirectory()
	if configDirError != nil {
		return "", configDirError
	}

	cachedTokenFile = filepath.Join(configDir, "token")
	return cachedTokenFile, nil
}

// Loads token stored in the filesystem and returns it.
func LoadToken() (string, error) {
    tokenFilePath, tokenError := GetTokenFile()
	if tokenError != nil {
		return "", tokenError
	}

    if _, readError := os.Stat(tokenFilePath); os.IsNotExist(readError) {
        os.Create(tokenFilePath)
		return Token, nil
    }

	tokenFile, openError := os.Open(tokenFilePath)

	if os.IsNotExist(openError) {
		return Token, nil
	}

	if openError != nil {
		return "", openError
	}

	defer tokenFile.Close()
	bToken, tokenLoadError := ioutil.ReadFile(tokenFilePath)

	//io.EOF would mean empty, therefore we use defaults.
	if tokenLoadError != nil {
		return "", tokenLoadError
	}

	return string(bToken), nil
}

func PersistToken() error {
	tokenFilePath, tokenError := GetTokenFile()
	if tokenError != nil {
		return tokenError
	}

	writeError := ioutil.WriteFile(tokenFilePath, []byte(Token), 0666)
	if writeError != nil {
		return writeError
	}

	return nil
}
