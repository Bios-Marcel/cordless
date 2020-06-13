package config

import (
	"encoding/json"
	"github.com/Bios-Marcel/cordless/util/files"
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
		ShortenWithExtension:                   false,
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
	//Token is the authorization token for accessing the discord API.
	Token string

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
	// ShortenWithExtension defines wether the suffix is added to the shortened
	// url
	ShortenWithExtension bool
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
	// ShowBottomBar decides whether an informational line is shown at the
	// bottom of cordless or not.
	ShowBottomBar bool

	// The image viewer to open when the user uses the "view attached images"
	// shortcut. This program will be passed a list of 1 or more image links
	// as if it were called from the command line, so the selected program
	// must be capable of opening image links.
	ImageViewer string
}

// Account has a name and a token. The name is just for the users recognition.
// The token is the actual token used to authenticate against the discord API.
type Account struct {
	Name  string
	Token string
}

var cachedConfigDir string
var cachedConfigFile string
var cachedScriptDir string

// SetConfigFile sets the config file path cache to the
// entered value
func SetConfigFile(configFilePath string) error {
	// get parent directory of config file
	parent := filepath.Dir(configFilePath)
	err := ensureDirectory(parent)
	if err == nil {
		cachedConfigFile = configFilePath
	} else {
		absolute, err := getAbsolutePath(parent)
		if err != nil {
			return err
		}
		err = ensureDirectory(absolute)
		if err == nil {
			cachedConfigFile = configFilePath
		}
	}
	return err
}

// GetConfigFile retrieves the config file path from cache
// or sets it to the default config file location
func GetConfigFile() (string, error) {
	if cachedConfigFile != "" {
		return cachedConfigFile, nil
	}

	configDir, configError := GetConfigDirectory()
	if configError != nil {
		return "", configError
	}

	cachedConfigFile = filepath.Join(configDir, "config.json")
	return cachedConfigFile, nil
}

// SetScriptDirectory sets the script directory cache
// to the specified value
func SetScriptDirectory(directoryPath string) error {
	err := ensureDirectory(directoryPath)
	if err == nil {
		cachedScriptDir = directoryPath
	} else {
		absolute, err := getAbsolutePath(directoryPath)
		if err != nil {
			return err
		}
		err = ensureDirectory(absolute)
		if err == nil {
			cachedConfigFile = absolute
		}
	}
	return err
}

// GetScriptDirectory retrieves the script path from cache
// or sets it to the default script directory location
func GetScriptDirectory() string {
	if cachedScriptDir != "" {
		return cachedScriptDir
	}
	cachedScriptDir = filepath.Join(cachedConfigDir, "scripts")
	return cachedScriptDir
}

// SetConfigDirectory sets the directory cache
func SetConfigDirectory(directoryPath string) error {
	err := ensureDirectory(directoryPath)
	if err == nil {
		cachedConfigDir = directoryPath
	} else {
		absolute, err := getAbsolutePath(directoryPath)
		if err != nil {
			return err
		}
		err = ensureDirectory(absolute)
		if err == nil {
			cachedConfigFile = absolute
		}
	}
	return err
}

// GetConfigDirectory retrieves the directory that stores
// cordless' settings from cache or sets it to the default
// location
func GetConfigDirectory() (string, error) {
	if cachedConfigDir != "" {
		return cachedConfigDir, nil
	}

	directory, err := getDefaultConfigDirectory()
	if err != nil {
		return "", err
	}

	statError := ensureDirectory(directory)
	if statError != nil {
		return "", statError
	}

	//After first retrieval, we will save this, as we needn't redo all that
	//stuff over and over again.
	cachedConfigDir = directory
	return cachedConfigDir, nil
}

func ensureDirectory(directoryPath string) error {
	_, statError := os.Stat(directoryPath)
	if os.IsNotExist(statError) {
		createDirsError := os.MkdirAll(directoryPath, 0766)
		if createDirsError != nil {
			return createDirsError
		}
		return nil
	}
	return statError
}

func getAbsolutePath(directoryPath string) (string, error) {
	absolutePath, resolveError := files.ToAbsolutePath(directoryPath)
	if resolveError != nil {
		return "", resolveError
	}
	return absolutePath, resolveError
}

//LoadConfig loads the configuration initially and returns it.
func LoadConfig() (*Config, error) {
	configFilePath, configError := GetConfigFile()
	if configError != nil {
		return nil, configError
	}

	configFile, openError := os.Open(configFilePath)

	if os.IsNotExist(openError) {
		return Current, nil
	}

	if openError != nil {
		return nil, openError
	}

	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	configLoadError := decoder.Decode(Current)

	//io.EOF would mean empty, therefore we use defaults.
	if configLoadError != nil && configLoadError != io.EOF {
		return nil, configLoadError
	}

	return Current, nil
}

// UpdateCurrentToken updates the current token and all accounts where the
// token was also used.
func UpdateCurrentToken(newToken string) {
	oldToken := Current.Token
	Current.Token = newToken
	for _, account := range Current.Accounts {
		if account.Token == oldToken {
			account.Token = newToken
		}
	}
}

//PersistConfig saves the current configuration onto the filesystem.
func PersistConfig() error {
	configFilePath, configError := GetConfigFile()
	if configError != nil {
		return configError
	}

	configAsJSON, jsonError := json.MarshalIndent(Current, "", "    ")
	if jsonError != nil {
		return jsonError
	}

	writeError := ioutil.WriteFile(configFilePath, configAsJSON, 0666)
	if writeError != nil {
		return writeError
	}

	return nil
}
