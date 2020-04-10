package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type AccountsFile struct {
	ActiveToken string
	Accounts    []*Account
}

// Account has a name and a token. The name is just for the users recognition.
// The token is the actual token used to authenticate against the discord API.
type Account struct {
	Name  string
	Token string
}

var LoadedAccountsFile *AccountsFile
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

func UpdateCurrentToken(newToken string) error {
	oldToken := LoadedAccountsFile.ActiveToken
	LoadedAccountsFile.ActiveToken = newToken
	for _, account := range LoadedAccountsFile.Accounts {
		if account.Token == oldToken {
			account.Token = newToken
		}
	}
	LoadedAccountsFile.ActiveToken = newToken
	persistError := PersistAccounts()
	return persistError
}

// LoadConfig loads the configuration initially and returns it.
func LoadAccounts() (*AccountsFile, error) {
	accountsFilePath, configError := GetAccountsFile()
	if configError != nil {
		return nil, configError
	}

	if _, readError := os.Stat(accountsFilePath); os.IsNotExist(readError) {
		ioutil.WriteFile(accountsFilePath, []byte("{\"ActiveToken\": \"\", \"Accounts\": []}"), 0600)
		return LoadAccounts()
	}

	accountsFile, openError := os.Open(accountsFilePath)

	if os.IsNotExist(openError) {
		return LoadedAccountsFile, nil
	}

	if openError != nil {
		return nil, openError
	}

	defer accountsFile.Close()
	decoder := json.NewDecoder(accountsFile)
	accountsLoadError := decoder.Decode(&LoadedAccountsFile)

	//io.EOF would mean empty, therefore we use defaults.
	if accountsLoadError != nil && accountsLoadError != io.EOF {
		return nil, accountsLoadError
	}

	return LoadedAccountsFile, nil
}

//PersistAccounts saves the current configuration onto the filesystem.
func PersistAccounts() error {
	accountsFilePath, accountsError := GetAccountsFile()
	if accountsError != nil {
		return accountsError
	}

	accountsAsJSON, jsonError := json.MarshalIndent(LoadedAccountsFile, "", "    ")
	if jsonError != nil {
		return jsonError
	}

	writeError := ioutil.WriteFile(accountsFilePath, accountsAsJSON, 0666)
	if writeError != nil {
		return writeError
	}

	return nil
}
