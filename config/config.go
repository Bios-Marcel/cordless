package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// AppName is the name representing the application.
	AppName = "Cordless"
	// AppNameLowercase is the representative name, but lowercase.
	// It us used for filepaths and such.
	AppNameLowercase = "cordless"

	// HourMinuteAndSeconds is the time format HH:MM:SS.
	HourMinuteAndSeconds = 0
	// HourAndMinute is the time format HH:MM.
	HourAndMinute = 1
	// NoTime means that not time at all will be displayed.
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
	currentConfig = &Config{
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
	}
)

// Config contains all possible configuration for the application.
type Config struct {
	// Token is the authorization token for accessing the discord API.
	Token string

	// Times decides on the time format (none, short and long).
	Times int
	// UseRandomUserColors decides whether the users get assigned a random color
	// out of a pool for the current session.
	UseRandomUserColors bool

	// FocusChannelAfterGuildSelection will cause the widget focus to move over
	// to the channel tree after selecting a guild.
	FocusChannelAfterGuildSelection bool
	// FocusMessageInputAfterChannelSelection will cause the widget focus to
	// move over to the message input widget after channel selection
	FocusMessageInputAfterChannelSelection bool

	// ShowUserContainer decides whether the user container is part of the
	// layout or not.
	ShowUserContainer bool
	// UseFixedLayout defines whether the FixedSizeLeft and FixedSizeRight
	// values will be applied or not.
	UseFixedLayout bool
	// FixedSizeLeft determines the size of the guilds/channels/friends
	// container on the left side of the layout.
	FixedSizeLeft int
	// FixedSizeRight defines the size of the users container on the right.
	FixedSizeRight int

	// OnTypeInListBehaviour defines whether the application focus the input
	// input field on typing, searches the list or does nothing.
	OnTypeInListBehaviour int
	// MouseEnabled decides whether the mouse is useable or not.
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

	// Accounts contains all saved accounts, allowing the user to dynamicly
	// switch between the accounts.
	Accounts []*Account

	// Show a padlock prefix of the channels that have access restriction
	IndicateChannelAccessRestriction bool
}

// Account has a name and a token. The name is just for the users recognition.
// The token is the actual token used to authenticate against the discord API.
type Account struct {
	Name  string
	Token string
}

var cachedConfigDir string
var cachedScriptDir string
var cachedConfigFilePath string

// SetConfigFile sets the configFileName cache to the entered value,
// otherwise cordless assumes default directory
func SetConfigFile(configFilePath string) (string, error) {
	parentDirectory := filepath.Dir(configFilePath)
	checkConfig := checkConfigDirectory(parentDirectory)
	if checkConfig != nil {
		return "", checkConfig
	}
	cachedConfigFilePath = configFilePath
	return cachedConfigFilePath, nil
}

//GetConfigFile returns the absolute path to the configuration file or an error
//in case of failure.
func GetConfigFile() (string, error) {
	// Prevents unnecessary overrides to
	// cachedConfigFilePath
	if cachedConfigFilePath != "" {
		return cachedConfigFilePath, nil
	}
	// Default behavior of configuration file under
	// app dir
	configDir, configError := GetConfigDirectory()
	if configError != nil {
		return "", configError
	}
	// Default config file is config.json
	return filepath.Join(configDir, "config.json"), nil
}

// GetScriptDirectory returns the path at which all the external scripts should
// lie.
func GetScriptDirectory() string {
	if cachedScriptDir == "" {
		cachedScriptDir = filepath.Join(cachedConfigDir, "scripts")
	}
	return cachedScriptDir
}

// SetConfigFile sets the configDirectory cache to the entered value,
// bypassing how cordless sets defaults.
func SetConfigDirectory(configPath string) (string, error) {
	checkConfig := checkConfigDirectory(configPath)
	if checkConfig != nil {
		return "", checkConfig
	}
	cachedConfigDir = configPath
	return cachedConfigDir, nil
}

// GetConfigDirectory is the parent directory in the os, that contains the
// settings for the application. If no directory is specified in cachedConfig
// it will assume cordless default parent directory
func GetConfigDirectory() (string, error) {
	// We don't want it overriding cache if there's been a directory
	// specified.
	if cachedConfigDir != "" {
		return cachedConfigDir, nil
	}

	directory, err := getDefaultConfigDirectory()
	if err != nil {
		return "", err
	}

	checkConfig := checkConfigDirectory(directory)
	if checkConfig != nil {
		return "", checkConfig
	}

	// Set cache so that this doesn't run over again
	cachedConfigDir = directory
	return cachedConfigDir, nil
}

// checkConfig handles errors that cordless can handle such as a directory not being
// there. If we get permission errors it is up to the user to fix it as there is nothing
// we can do.
func checkConfigDirectory(directoryPath string) error {
	_, statError := os.Stat(directoryPath)
	if os.IsNotExist(statError) {
		// Folders have to be executable, hence 766 instead of 666.
		createDirsError := os.MkdirAll(directoryPath, 0766)
		if createDirsError != nil {
			return createDirsError
		}
	}
	return statError
}

// GetConfig returns the currently loaded config object
func GetConfig() *Config {
	return currentConfig
}

// LoadConfig loads the configuration initially and returns it.
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
	configLoadError := decoder.Decode(currentConfig)

	// io.EOF would mean empty, therefore we use defaults.
	if configLoadError != nil && configLoadError != io.EOF {
		return nil, configLoadError
	}

	return GetConfig(), nil
}

// UpdateCurrentToken updates the current token and all accounts where the
// token was also used.
func UpdateCurrentToken(newToken string) {
	oldToken := currentConfig.Token
	currentConfig.Token = newToken
	for _, account := range currentConfig.Accounts {
		if account.Token == oldToken {
			account.Token = newToken
		}
	}
}

// PersistConfig saves the current configuration onto the filesystem.
func PersistConfig() error {
	configFilePath, configError := GetConfigFile()
	if configError != nil {
		return configError
	}

	configAsJSON, jsonError := json.MarshalIndent(currentConfig, "", "    ")
	if jsonError != nil {
		return jsonError
	}

	writeError := ioutil.WriteFile(configFilePath, configAsJSON, 0666)
	if writeError != nil {
		return writeError
	}

	return nil
}
