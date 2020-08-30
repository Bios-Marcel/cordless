package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bios-Marcel/cordless/util/files"
)

const (
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

	// NoColor will render all usernames using the PrimaryTextColor defined in
	// the theme that's currently loaded.
	NoColor UserColor = "none"
	// SingleColor causes cordless to take the color specified in the theme
	SingleColor UserColor = "single"
	// RandomColor causes cordless to take a random color from the theme
	// specified pool of usable random colors.
	RandomColor UserColor = "random"
	// RoleColor attempts to use the first colored role it finds for a user.
	RoleColor UserColor = "role"
)

// UserColor represents available configurations for rendering a users color.
type UserColor string

var (
	//Current is the currently loaded configuration.
	Current = createDefaultConfig()
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
	//UserColors decides how cordless determines in which color it displays
	//a user in the chat or the user tree.
	UserColors UserColor

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
	// ShortenWithExtension defines whether the suffix is added to the shortened
	// url
	ShortenWithExtension bool
	// ShortenerPort defines the port, that the webserver for the linkshortener
	// will be using.
	ShortenerPort int

	// DesktopNotifications decides whether a popup will be shown in the users
	// system when a notification needs to be sent.
	DesktopNotifications bool

	// DesktopNotificationsUserInactivityThreshold defines how many seconds
	// have to pass between now and the last user input (key stroke) in order
	// for a notification to be sent.
	DesktopNotificationsUserInactivityThreshold int

	// DesktopNotificationsForLoadedChannel Defines whether notifications are
	// also sent for the currently selected (loaded) channel.
	DesktopNotificationsForLoadedChannel bool

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

	// Accounts contains all saved accounts, allowing the user to dynamically
	// switch between the accounts.
	Accounts []*Account

	// Show a padlock prefix of the channels that have access restriction
	IndicateChannelAccessRestriction bool
	// ShowBottomBar decides whether an informational line is shown at the
	// bottom of cordless or not.
	ShowBottomBar bool

	// ShowNicknames decides whether nicknames are shown throughout the
	// application, as there are some childish goons that deem it funny
	// to impersonate people or change their name every 5 minutes.
	ShowNicknames bool

	// FileHandlers allow registering specific file-handers for certain
	FileOpenHandlers map[string]string
	// FileOpenSaveFilesPermanently decides whether opened files are saved
	// in the system cache (temporary) or in the user specified path.
	FileOpenSaveFilesPermanently bool
	// FileDownloadSaveLocation defines the folder where cordless generally
	// download files to. If FileOpenSaveFilesPermanently has been set to
	// true, then all opened files are saved in this folder for example.
	FileDownloadSaveLocation string
}

// Account has a name and a token. The name is just for the users recognition.
// The token is the actual token used to authenticate against the discord API.
type Account struct {
	Name  string
	Token string
}

// GetAccountToken returns the token for the given account or an empty string
// if the account can't be found or has no token set up.
func (config *Config) GetAccountToken(account string) string {
	for _, acc := range config.Accounts {
		if strings.EqualFold(acc.Name, account) {
			return acc.Token
		}
	}

	return ""
}

var cachedConfigDir string
var cachedConfigFile string
var cachedScriptDir string

func createDefaultConfig() *Config {
	// The values here are the defaults which can / will be overwritten
	// by loading the config file. This can also be used to reset the
	// active set of settings during runtime.
	return &Config{
		Autocomplete:                           true,
		Times:                                  HourMinuteAndSeconds,
		UserColors:                             SingleColor,
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
		DesktopNotificationsUserInactivityThreshold: 10,
		DesktopNotificationsForLoadedChannel:        true,
		ShowPlaceholderForBlockedMessages:           true,
		DontShowUpdateNotificationFor:               "",
		ShowUpdateNotifications:                     true,
		IndicateChannelAccessRestriction:            false,
		ShowBottomBar:                               true,
		ShowNicknames:                               true,
		FileOpenHandlers:                            make(map[string]string),
		FileOpenSaveFilesPermanently:                false,
		FileDownloadSaveLocation:                    "~/Downloads",
	}
}

// SetConfigFile sets the config file path cache to the
// entered value
func SetConfigFile(configFilePath string) error {
	// get parent directory of config file
	parent := filepath.Dir(configFilePath)
	err := files.EnsureDirectoryExists(parent)
	if err == nil {
		cachedConfigFile = configFilePath
	} else {
		absolute, err := getAbsolutePath(parent)
		if err != nil {
			return err
		}
		err = files.EnsureDirectoryExists(absolute)
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
	err := files.EnsureDirectoryExists(directoryPath)
	if err == nil {
		cachedScriptDir = directoryPath
	} else {
		absolute, err := getAbsolutePath(directoryPath)
		if err != nil {
			return err
		}
		err = files.EnsureDirectoryExists(absolute)
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
	err := files.EnsureDirectoryExists(directoryPath)
	if err == nil {
		cachedConfigDir = directoryPath
	} else {
		absolute, err := getAbsolutePath(directoryPath)
		if err != nil {
			return err
		}
		err = files.EnsureDirectoryExists(absolute)
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

	statError := files.EnsureDirectoryExists(directory)
	if statError != nil {
		return "", statError
	}

	//After first retrieval, we will save this, as we needn't redo all that
	//stuff over and over again.
	cachedConfigDir = directory
	return cachedConfigDir, nil
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
