package shortcuts

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/gdamore/tcell"
)

var (
	globalScope = &Scope{
		Parent:     nil,
		Identifier: "global",
		Name:       "Application wide",
	}

	scopes = []*Scope{
		globalScope,
	}

	FocusChannelContainer = &Shortcut{
		Name:       "Focus channel container",
		Identifier: "focus_channel_container",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModAlt),
	}

	FocusUserContainer = &Shortcut{
		Name:       "Focus user container",
		Identifier: "focus_user_container",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModAlt),
	}

	ToggleUserContainer = &Shortcut{
		Name:       "Toggle user container",
		Identifier: "toggle_user_container",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyRune, 'U', tcell.ModAlt),
	}

	FocusGuildContainer = &Shortcut{
		Name:       "Focus guild container",
		Identifier: "focus_guild_container",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModAlt),
	}

	FocusPrivateChatPage = &Shortcut{
		Name:       "Focus private chat page",
		Identifier: "focus_private_chat_page",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModAlt),
	}

	FocusMessageInput = &Shortcut{
		Name:       "Focus message input",
		Identifier: "focus_message_input",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModAlt),
	}

	FocusMessageContainer = &Shortcut{
		Name:       "Focus message container",
		Identifier: "focus_message_container",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModAlt),
	}

	FocusCommandInput = &Shortcut{
		Name:       "Focus command input",
		Identifier: "focus_command_input",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyCtrlI, rune(tcell.KeyCtrlI), tcell.ModNone),
	}

	FocusCommandOutput = &Shortcut{
		Name:       "Focus command output",
		Identifier: "focus_command_output",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyCtrlO, rune(tcell.KeyCtrlO), tcell.ModCtrl),
	}

	ToggleCommandView = &Shortcut{
		Name:       "Toggle command view",
		Identifier: "toggle_command_view",
		scope:      globalScope,
		Event:      tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModAlt),
	}

	Shortcuts = []*Shortcut{
		FocusChannelContainer,
		FocusUserContainer,
		ToggleUserContainer,
		FocusGuildContainer,
		FocusPrivateChatPage,
		FocusMessageInput,
		FocusMessageContainer,
		FocusCommandInput,
		FocusCommandOutput,
		ToggleCommandView,
	}
)

// Shortcut defines a shortcut within the application. The scope might for
// example be a widget or situation in which the user is.
type Shortcut struct {
	// Identifier will be used for persistence and should never change
	Identifier string

	// Name will be shown on the UI
	Name string

	// The scope will be omitted, as this needed be persisted anyway.
	scope *Scope

	// Event is the shortcut expressed as it's resulting tcell Event.
	Event *tcell.EventKey
}

// Scope is what describes a shortcuts scope within the application. Usually
// a scope can only have a specific shortcut once and a children scope will
// overwrite that shortcut, since that lower scope has the upper hand.
type Scope struct {
	// Parent is this scopes upper Scope, which may be null, in case this is a
	// root scope.
	Parent *Scope

	// Identifier will be used for persistence and should never change
	Identifier string

	// Name will be shown on the UI
	Name string
}

// ShortcutDataRepresentation represents a shortcut on the users harddrive.
// This prevents redudancy of scopes.
type ShortcutDataRepresentation struct {
	Identifier      string
	ScopeIdentifier string
	EventKey        tcell.Key
	EventRune       rune
	EventMod        tcell.ModMask
}

// MarshalJSON marshals a Shortcut into a ShortcutDataRepresentation. This
// happens in order to prevent saving the scopes multiple times, therefore
// the scopes will be hardcoded.
func (shortcut *Shortcut) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(&ShortcutDataRepresentation{
		Identifier:      shortcut.Identifier,
		ScopeIdentifier: shortcut.scope.Identifier,
		EventKey:        shortcut.Event.Key(),
		EventRune:       shortcut.Event.Rune(),
		EventMod:        shortcut.Event.Modifiers(),
	}, "", "    ")
}

// UnmarshalJSON unmarshals ShortcutDataRepresentation JSON into a Shortcut.
// This happens in order to prevent saving the scopes multiple times,
// therefore the scopes will be hardcoded.
func (shortcut *Shortcut) UnmarshalJSON(data []byte) error {
	var temp ShortcutDataRepresentation
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	shortcut.Event = tcell.NewEventKey(temp.EventKey, temp.EventRune, temp.EventMod)
	shortcut.Identifier = temp.Identifier

	for _, scope := range scopes {
		if scope.Identifier == temp.ScopeIdentifier {
			shortcut.scope = scope
			return nil
		}
	}

	return fmt.Errorf(fmt.Sprintf("error finding scope '%s'", temp.ScopeIdentifier))
}

func getShortcutsPath() (string, error) {
	configDirecotry, configError := config.GetConfigDirectory()
	if configError != nil {
		return "", configError
	}

	return filepath.Join(configDirecotry, "shortcuts.json"), nil
}

// Load loads the shortcuts and copies the events to the correct shortcuts
// that reside inside of the memory.
func Load() error {
	shortcutsPath, pathError := getShortcutsPath()
	if pathError != nil {
		return pathError
	}

	shortcutsFile, openError := os.Open(shortcutsPath)

	if os.IsNotExist(openError) {
		return nil
	}

	if openError != nil {
		return openError
	}

	defer shortcutsFile.Close()
	decoder := json.NewDecoder(shortcutsFile)
	tempShortcuts := make([]*Shortcut, 0)
	shortcutsLoadError := decoder.Decode(&tempShortcuts)

	//io.EOF would mean empty, therefore we use defaults.
	if shortcutsLoadError != nil && shortcutsLoadError != io.EOF {
		return nil
	}

	if shortcutsLoadError != nil {
		return shortcutsLoadError
	}

OUTER_LOOP:
	for _, shortcut := range tempShortcuts {
		for _, otherShortcut := range Shortcuts {
			if otherShortcut.Identifier == shortcut.Identifier &&
				otherShortcut.scope.Identifier == shortcut.scope.Identifier {
				otherShortcut.Event = shortcut.Event
				continue OUTER_LOOP
			}
		}
	}

	return nil
}

// Persist saves the currently held in memory held shortcuts
func Persist() error {
	filePath, pathError := getShortcutsPath()
	if pathError != nil {
		return pathError
	}

	shortcutsAsJSON, jsonError := json.MarshalIndent(&Shortcuts, "", "    ")
	if jsonError != nil {
		return jsonError
	}

	writeError := ioutil.WriteFile(filePath, shortcutsAsJSON, 0666)
	if writeError != nil {
		return writeError
	}

	return nil
}
