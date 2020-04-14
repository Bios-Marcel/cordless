package config

import (
	"fmt"
	"path/filepath"

	"github.com/Bios-Marcel/cordless/util/files"
)

const (
	//AppName is the name representing the application.
	AppName = "Cordless"
	//AppNameLowercase is the representative name, but lowercase.
	//It us used for filepaths and such.
	AppNameLowercase = "cordless"

	//HourMinuteAndSeconds is the time format HH:MM:SS.
	HourMinuteAndSeconds = 0
	//HourAndMinute is the time format HH:MM.
	HourAndMinute = 1
	//NoTime means that not time at all will be displayed.
	NoTime = 2

	// DoNothingOnTypeInList means that when typing in a list (treeview) simply
	// nothing will happen.
	DoNothingOnTypeInList = 0
	// SearchOnTypeInList will cause the widget to search through it's
	// children, trying to find anything that is prefixed with the
	// previously entered characters.
	SearchOnTypeInList = 1
	// FocusMessageInputOnTypeInList will automatically focus the message input
	// component and transfer the typed character into it as well.
	FocusMessageInputOnTypeInList = 2
)

var (
	Current = &Config{
		Autocomplete:                           true,
		Times:                                  HourMinuteAndSeconds,
		UseRandomUserColors:                    false,
		ShowUserContainer:                      true,
		UseFixedLayout:                         false,
		FixedSizeLeft:                          12,
		FixedSizeRight:                         12,
		FocusChannelAfterGuildSelection:        true,
		FocusMessageInputAfterChannelSelection: true,
		OnTypeInListBehaviour:                  SearchOnTypeInList,
		MouseEnabled:                           true,
		ShortenLinks:                           false,
		ShortenerPort:                          63212,
		DesktopNotifications:                   true,
		ShowPlaceholderForBlockedMessages:      true,
		DontShowUpdateNotificationFor:          "",
		ShowUpdateNotifications:                true,
		IndicateChannelAccessRestriction:       false,
		ShowBottomBar:                          true,
		ImageViewer:                            "feh",
	}
)

//Config contains all possible configuration for the application.
type Config struct {
	//Autocomplete decides whether the chat automatically offers autocomplete
	//values for the currently given text.
	Autocomplete bool

	//Times decides on the time format (none, short and long).
	Times int
	//UseRandomUserColors decides whether the users get assigned a random color
	//out of a pool for the current session.
	UseRandomUserColors bool

	//FocusChannelAfterGuildSelection will cause the widget focus to move over
	//to the channel tree after selecting a guild.
	FocusChannelAfterGuildSelection bool
	//FocusMessageInputAfterChannelSelection will cause the widget focus to
	//move over to the message input widget after channel selection
	FocusMessageInputAfterChannelSelection bool

	//ShowUserContainer decides whether the user container is part of the
	//layout or not.
	ShowUserContainer bool
	//UseFixedLayout defines whether the FixedSizeLeft and FixedSizeRight
	//values will be applied or not.
	UseFixedLayout bool
	//FixedSizeLeft determines the size of the guilds/channels/friends
	//container on the left side of the layout.
	FixedSizeLeft int
	//FixedSizeRight defines the size of the users container on the right.
	FixedSizeRight int

	// OnTypeInListBehaviour defines whether the application focus the input
	// input field on typing, searches the list or does nothing.
	OnTypeInListBehaviour int
	// MouseEnabled decides whether the mouse is usable or not.
	MouseEnabled bool

	// ShortenLinks decides whether cordless starts a local webserver in order
	// to be able to shorten links
	ShortenLinks bool
	// ShortenerPort defines the port, that the webserver for the linkshortener
	// will be using.
	ShortenerPort int

	// DesktopNotifications decides whether a popup will be shown in the users
	// system when a notification needs to be sent.
	DesktopNotifications bool

	// ShowPlaceholderForBlockedMessages will cause blocked message to shown
	// as a placeholder message, replacing user and message with generic text.
	// The time of the message will still be correct in order to not mess up
	// the timeline of messages.
	ShowPlaceholderForBlockedMessages bool

	// ShowUpdateNotifications decides whether update notifications are
	// shown at all.
	ShowUpdateNotifications bool
	// DontShowUpdateNotificationFor decides what version to skip update
	// notifications for. Since there's always only one latest version, this
	// is a string and not an array of strings.
	DontShowUpdateNotificationFor string

	// Show a padlock prefix of the channels that have access restriction
	IndicateChannelAccessRestriction bool
	// ShowBottomBar decides whether an informational line is shown at the
	// bottom of cordless or not.
	ShowBottomBar bool

	// The image viewer to open when the user uses the "view attached images"
	// shortcut. This program will be passed a list of 1 or more image links
	// as if it were called from the command line, so the selected program
	// must be capable of opening image links.
	ImageViewer string
}

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

var LoadedAccountsFile = &AccountsFile{
    ActiveToken: "",
    Accounts: []*Account{},
}

var cachedAccountsFile string
var cachedConfigFile string
var cachedConfigDir string
var cachedScriptDir string

// SetConfigFile sets the config file path cache to the entered value.
func SetConfigFile(path string) error {
    existsError := files.CheckExists(path)
    if existsError == nil {
	    cachedConfigFile = path
        return nil
    }
    return existsError
}

// SetAccountsFile sets the accounts file path cache to the entered value.
func SetAccountsFile(path string) error {
    existsError := files.CheckExists(path)
    if existsError == nil {
	    cachedAccountsFile = path
        return nil
    }
    return existsError
}

func DefaultConfigTarget(targetName string) (string, error) {
	configDir, configError := GetConfigDirectory()
	if configError != nil {
		return "", configError
	}

    absPath := filepath.Join(configDir, targetName)
	return absPath, nil
}

// GetConfigFile retrieves the config file path from cache
// or sets it to the default config file location
func GetConfigFile() (string, error) {
	if cachedConfigFile != "" {
		return cachedConfigFile, nil
	}
    defaultTarget, defaultTargetError := DefaultConfigTarget("config.json")
    if defaultTargetError != nil {
        return "", defaultTargetError
    }
    cachedConfigFile = defaultTarget
    return defaultTarget, nil

}

// GetConfigFile retrieves the config file path from cache or sets it to the
// default config file location
func GetAccountsFile() (string, error) {
	if cachedAccountsFile != "" {
		return cachedAccountsFile, nil
	}
    defaultTarget, defaultTargetError := DefaultConfigTarget("accounts.json")
    if defaultTargetError != nil {
        return "", defaultTargetError
    }
    cachedAccountsFile = defaultTarget
    return defaultTarget, nil

}

// SetScriptDirectory sets the script directory cache to the specified value.
func SetScriptDirectory(path string) error {
    existsError := files.CheckExists(path)
    if existsError == nil {
	    cachedScriptDir = path
        return nil
    }
    return existsError
}

// GetScriptDirectory retrieves the script path from cache or sets it to the
// default script directory location.
func GetScriptDirectory() string {
    // TODO This function is not implemented with error checking anywhere. Until
    // then this function will not return an error, unlike the simmilar
    // functions.
	if cachedScriptDir != "" {
		return cachedScriptDir
	}
	cachedScriptDir = filepath.Join(cachedConfigDir, "scripts")
	return cachedScriptDir
}

// SetScriptDirectory sets the config directory cache to the specified value.
func SetConfigDirectory(path string) error {
	existsError := files.CheckExists(path)
    if existsError == nil {
	    cachedConfigDir = path
        return nil
    }
    return existsError
}

// GetConfigDirectory retrieves the directory that stores cordless' settings
// from cache or sets it to the default location.
func GetConfigDirectory() (string, error) {
	if cachedConfigDir != "" {
		return cachedConfigDir, nil
	}

	directory, err := getDefaultConfigDirectory()
	if err != nil {
		return "", err
	}

	statError := files.EnsureDirectory(directory)
	if statError != nil {
		return "", statError
	}

	//After first retrieval, we will save this, as we needn't redo all that
	//stuff over and over again.
	cachedConfigDir = directory
	return cachedConfigDir, nil
}


// LoadConfig loads the configuration initially and returns it.
func LoadConfig() (*Config, error) {
	configFilePath, configError := GetConfigFile()
	if configError != nil {
		return nil, configError
	}
    fmt.Printf(configFilePath)
    if files.CheckExists(configFilePath) != nil {
        persistsError := PersistConfig()
        if persistsError != nil {
            return nil, persistsError
        }
        return LoadConfig()
    }
    jsonError := files.LoadJSON(configFilePath, Current)
    if jsonError != nil {
        return nil, jsonError
    }
    return Current, nil
}

func LoadAccounts() (*AccountsFile, error) {
	accountsFilePath, accountsError := GetAccountsFile()
	if accountsError!= nil {
		return nil, accountsError
	}
    if files.CheckExists(accountsFilePath) != nil {
        persistsError := PersistConfig()
        if persistsError != nil {
            return nil, persistsError
        }
        return LoadAccounts()
    }
    jsonError := files.LoadJSON(accountsFilePath, LoadedAccountsFile)
    if jsonError != nil {
        return nil, jsonError
    }
    return LoadedAccountsFile, nil
}

// PersistConfig saves the current configuration onto the filesystem.
func PersistConfig() error {
	configFilePath, configError := GetConfigFile()
	if configError != nil {
		return configError
	}
    return files.WriteJSON(configFilePath, Current)
}

// PersistConfig saves the current account configuration onto the
// filesystem.
func PersistAccounts() error {
	accountsFilePath, accountsError := GetAccountsFile()
	if accountsError!= nil {
		return accountsError
	}
    return files.WriteJSON(accountsFilePath, LoadedAccountsFile)
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

