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
	currentConfig = Config{
		Times:                                  HourMinuteAndSeconds,
		ShowUserContainer:                      true,
		ShowChatHeader:                         true,
		ShowFrame:                              false,
		UseFixedLayout:                         false,
		FixedSizeLeft:                          12,
		FixedSizeRight:                         12,
		FocusChannelAfterGuildSelection:        true,
		FocusMessageInputAfterChannelSelection: true,
		OnTypeInListBehaviour:                  SearchOnTypeInList,
		ShortenLinks:                           false,
		ShortenerPort:                          63212,
		DesktopNotifications:                   true,
	}
)

//Config contains all possible configuration for the application.
type Config struct {
	//Token is the authorization token for accessing the discord API.
	Token string

	//Times decides on the time format (none, short and long).
	Times int

	//FocusChannelAfterGuildSelection will cause the widget focus to move over
	//to the channel tree after selecting a guild.
	FocusChannelAfterGuildSelection bool
	//FocusMessageInputAfterChannelSelection will cause the widget focus to
	//move over to the message input widget after channel selection
	FocusMessageInputAfterChannelSelection bool

	//ShowFrame decides wether the application will have a border and a title.
	ShowFrame bool
	//ShowChatHeader decides wether the box above the chat will be shown
	//or not.
	ShowChatHeader bool
	//ShowUserContainer decides wether the user container is part of the
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

	// OnTypeInListBehaviour defines wether the application focus the input
	// input field on typing, searches the list or does nothing.
	OnTypeInListBehaviour int

	// ShortenLinks decides wether cordless starts a local webserver in order
	// to be able to shorten links
	ShortenLinks bool
	// ShortenerPort defines the port, that the webserver for the linkshortener
	// will be using.
	ShortenerPort int

	// DesktopNotifications decides wether a popup will be shown in the users
	// system when a notification needs to be sent.
	DesktopNotifications bool

	// Accounts contains all saved accounts, allowing the user to dynamicly
	// switch between the accounts.
	Accounts []*Account
}

// Account has a name and a token. The name is just for the users recognition.
// The token is the actual token used to authenticate against the discord API.
type Account struct {
	Name  string
	Token string
}

var cachedConfigDir string
var cachedScriptDir string

//GetConfigFile returns the absolute path to the configuration file or an error
//in case of failure.
func GetConfigFile() (string, error) {
	configDir, configError := GetConfigDirectory()

	if configError != nil {
		return "", configError
	}

	return filepath.Join(configDir, "config.json"), nil
}

//GetScriptDirectory returns the path at which all the external scripts should
//lie.
func GetScriptDirectory() string {
	if cachedScriptDir == "" {
		//We'll just make the assumption, that the config dir has already been
		//initialized at that point and time in the application.
		cachedScriptDir = filepath.Join(cachedConfigDir, "scripts")
	}
	return cachedScriptDir
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
