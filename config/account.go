package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Account has a name and a token. The name is just for the users recognition.
// The token is the actual token used to authenticate against the discord API.
type Account struct {
	Name  string
	Token string
}

var ActiveAccount *Account
var Accounts []*Account
var cachedAccountsFile string

// SetAccountsFile sets the accounts file path cache to the
// entered value
func SetAccountsFile(accountsFilePath string) error {
	// get parent directory of accounts file
	parent := filepath.Dir(accountsFilePath)
	err := ensureDirectory(parent)
	if err == nil {
		cachedAccountsFile = accountsFilePath
	} else {
		absolute, err := getAbsolutePath(parent)
		if err != nil {
			return err
		}
		err = ensureDirectory(absolute)
		if err == nil {
			cachedAccountsFile = accountsFilePath
		}
	}
	return err
}

// GetTokenFile retrieves the config file path from cache
// or sets it to the default config file location
func GetAccountsFile() (string, error) {
	if cachedAccountsFile != "" {
		return cachedAccountsFile, nil
	}

	configDir, configDirError := GetConfigDirectory()
	if configDirError != nil {
		return "", configDirError
	}

	cachedAccountsFile = filepath.Join(configDir, "accounts.json")
	return cachedAccountsFile, nil
}

// LoadConfig loads the configuration initially and returns it.
func LoadAccounts() ([]*Account, error) {
	accountsFilePath, configError := GetAccountsFile()
	if configError != nil {
		return nil, configError
	}

    if _, readError := os.Stat(accountsFilePath); os.IsNotExist(readError) {
        os.Create(accountsFilePath)
		return Accounts, nil
    }


	accountsFile, openError := os.Open(accountsFilePath)

	if os.IsNotExist(openError) {
		return Accounts, nil
	}

	if openError != nil {
		return nil, openError
	}

	defer accountsFile.Close()
	decoder := json.NewDecoder(accountsFile)
	accountsLoadError := decoder.Decode(&Accounts)

	//io.EOF would mean empty, therefore we use defaults.
	if accountsLoadError != nil && accountsLoadError != io.EOF {
		return nil, accountsLoadError
	}

	return Accounts, nil
}

//PersistAccounts saves the current configuration onto the filesystem.
func PersistAccounts() error {
	accountsFilePath, accountsError := GetAccountsFile()
	if accountsError != nil {
		return accountsError
	}

	accountsAsJSON, jsonError := json.MarshalIndent(Accounts, "", "    ")
	if jsonError != nil {
		return jsonError
	}

	writeError := ioutil.WriteFile(accountsFilePath, accountsAsJSON, 0666)
	if writeError != nil {
		return writeError
	}

	return nil
}
